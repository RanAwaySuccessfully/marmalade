import mediapipe as mp
from mediapipe.tasks import python
from mediapipe.tasks.python import vision

from threading import Lock

from compute_params import (
    compute_params_from_blendshapes,
    compute_params_from_matrix,
)

import socket
import time
import cv2
import json
import time


class ResultTracker:
    def __init__(self, max_failures):
        self.lock = Lock()
        self.failures = 0
        self.max_failures = max_failures

    def add_failure(self):
        with self.lock:
            self.failures += 1

    def reset(self):
        with self.lock:
            self.failures = 0

    def is_disconnected(self):
        with self.lock:
            return self.failures > self.max_failures

def format_blendshapes(bs):
    return {
        "k": bs.category_name[0].upper() + bs.category_name[1:],
        "v": bs.score,
    }

def send_data(detection_result, timestamp, target):
    udp_target = target.split(":")
    ip = udp_target[0]
    port = int(udp_target[1])

    faceFound = False
    face_params = {"Rotation": {}, "Position": {}}
    eye_params = {"EyeLeft": {}, "EyeRight": {}}

    blendshapes = []

    face_blendshapes_list = detection_result.face_blendshapes
    if len(face_blendshapes_list) != 0:
        faceFound = True
        face_blendshapes = face_blendshapes_list[0]
        blendshapes = map(format_blendshapes, face_blendshapes)

        face_params = compute_params_from_matrix(detection_result.facial_transformation_matrixes[0])
        eye_params = compute_params_from_blendshapes(face_blendshapes)

    data = {
        "Timestamp": round(time.time()),
        "Hotkey": -1,
        "FaceFound": faceFound,
        "Rotation": face_params["Rotation"],
        "Position": face_params["Position"],
        "BlendShapes": list(blendshapes),
        "EyeLeft": eye_params["EyeLeft"],
        "EyeRight": eye_params["EyeRight"],
    }

    jsonstr = json.dumps(data)

    '''
    with open("debug.log", "a") as f:
        f.write(jsonstr)
        f.write("\n")
    '''

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.sendto(bytes(jsonstr, "utf-8"), (ip, port))


def main():
    # webcam reader
    # make these parameters?
    # camera_id = args.camera
    # width = args.width
    # height = args.height
    # fps = args.fps
    camera_id = 2
    width = 1280
    height = 720
    fps = 30
    use_gpu = False
    model = 'face_landmarker_v2_with_blendshapes.task'
    websocket_failures = 5
    camera_failures = 5
    target = ''

    capture = cv2.VideoCapture()
    capture.set(cv2.CAP_PROP_FRAME_WIDTH, width)
    capture.set(cv2.CAP_PROP_FRAME_HEIGHT, height)
    capture.set(cv2.CAP_PROP_FPS, fps)
    capture.open(camera_id)
    time.sleep(0.02)  # allow camera to initialize

    if capture.isOpened() == False:
        print("Device not opened")
        exit(1)

    attempts = 0
    result_tracker = ResultTracker(websocket_failures)

    def process_results(
        detection_result: mp.tasks.vision.FaceLandmarkerResult,
        image: mp.Image,
        timestamp_ms: int,
    ):
        send_data(detection_result, timestamp_ms, target)
        # if result != True:
        #     result_tracker.add_failure()
        #else:
        #    result_tracker.reset()

    delagate = python.BaseOptions.Delegate.CPU
    if use_gpu:
        delagate = python.BaseOptions.Delegate.GPU

    base_options = python.BaseOptions(
        model_asset_path=model, delegate=delagate
    )

    options = vision.FaceLandmarkerOptions(
        base_options,
        running_mode=mp.tasks.vision.RunningMode.LIVE_STREAM,
        output_face_blendshapes=True,
        output_facial_transformation_matrixes=True,
        num_faces=1,
        result_callback=process_results,
    )

    detector = vision.FaceLandmarker.create_from_options(options)
    fps = capture.get(cv2.CAP_PROP_FPS)
    wait_interval_sec = 0.1 / fps  # wait 10% of the time to get a frame

    try:
        while True:
            target = input()

            # Load image
            ret, cv2_image = capture.read()

            if result_tracker.is_disconnected():
                print("No longer recieving from server, disconnecting!")
                break

            if ret:
                attempts = 0
                image = mp.Image(image_format=mp.ImageFormat.SRGB, data=cv2_image)
                timestamp = int(capture.get(cv2.CAP_PROP_POS_MSEC))
                detector.detect_async(image, timestamp)
            else:
                attempts += 1
                time.sleep(wait_interval_sec)
            if attempts > camera_failures:
                print("Too many failed attempts getting camera image, quitting")
                break
    except KeyboardInterrupt:
        print("Quitting")
    
    capture.release()

main()