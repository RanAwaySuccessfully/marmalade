import mediapipe as mp
from mediapipe.tasks import python
from mediapipe.tasks.python import vision

import cv2 # opencv-python

from compute_params import (
    compute_params_from_blendshapes,
    compute_params_from_matrix,
)

from threading import Lock
import socket
import time
import json
import argparse
import signal
import sys
import threading


clients = set()

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

    key = bs.category_name[0].upper() + bs.category_name[1:]

    # i need to invert this for VTubing's sake
    isLeft = key[-4:] == "Left"
    isRight = key[-5:] == "Right"

    if isLeft:
        key = key.replace("Left", "Right")
    
    if isRight:
        key = key.replace("Right", "Left")

    return {
        "k": key,
        "v": bs.score,
    }

def send_data(detection_result, timestamp):
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
    success = True

    for client in clients:
        udp_client = client.split(":")
        ip = udp_client[0]
        port = int(udp_client[1])

        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            sock.sendto(bytes(jsonstr, "utf-8"), (ip, port))
        except socket.timeout:
            success = False
    
    return success


def get_args():
    parser = argparse.ArgumentParser(
        prog="lilacsMediaPipeForward",
        description="Plugin for VTube Studio that forwards blendshapes and transforms from google's mediapipe face landmarker",
    )
    parser.add_argument(
        "-m",
        "--model",
        help="mediapipe model file",
        default="face_landmarker_v2_with_blendshapes.task",
    )
    parser.add_argument("-c", "--camera", help="index of camera device", default=0)
    parser.add_argument("-W", "--width", help="width of camera image", default=1280)
    parser.add_argument("-H", "--height", help="height of camera image", default=720)
    parser.add_argument("-f", "--fps", help="frame rate of the camera", default=30)
    parser.add_argument("-t", "--fmt", help="encoding format", default='YUYV')
    parser.add_argument("-g", "--use-gpu", default=False, action="store_true")
    parser.add_argument(
        "--camera-failures",
        help="Number of failed frames to grab before quitting",
        default=5,
    )
    parser.add_argument(
        "--websocket-failures",
        help="Number of failed communications to vtube studio before quitting",
        default=5,
    )
    return parser.parse_args()

ended = False

def main(args):
    camera_id = int(args.camera)
    width = float(args.width)
    height = float(args.height)
    fps = float(args.fps)
    use_gpu = bool(args.use_gpu)
    model = args.model
    websocket_failures = float(args.websocket_failures)
    camera_failures = float(args.camera_failures)
    cam_format = args.fmt

    capture = cv2.VideoCapture(camera_id, cv2.CAP_V4L2)
    capture.set(cv2.CAP_PROP_FOURCC, cv2.VideoWriter_fourcc(*cam_format))
    capture.set(cv2.CAP_PROP_FRAME_WIDTH, width)
    capture.set(cv2.CAP_PROP_FRAME_HEIGHT, height)
    capture.set(cv2.CAP_PROP_FPS, fps)
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
        result = send_data(detection_result, timestamp_ms)
        if result != True:
            result_tracker.add_failure()
        else:
            result_tracker.reset()

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
    # fourcc = capture.get(cv2.CAP_PROP_FOURCC)
    # print(int(fourcc).to_bytes(4, byteorder=sys.byteorder).decode())
    fps = capture.get(cv2.CAP_PROP_FPS)
    wait_interval_sec = 0.1 / fps  # wait 10% of the time to get a frame

    try:
        while not ended:
            if len(clients) == 0:
                change = input()
                t1 = threading.Thread(target=client_update, args=(change,))
                t1.start()

            # start = time.time()
            # Load image
            ret, cv2_image = capture.read()
            # print((time.time() - start) * 1000)

            if result_tracker.is_disconnected():
                print("No longer recieving from server, disconnecting!")
                sys.exit(170)
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
                sys.exit(171)
                break
    except KeyboardInterrupt:
        print("Quitting")
    
    capture.release()

def client_update(first_change):
    change = first_change
    idle = False

    global clients
    while not idle:
        if change == "":
            change = input()

        if change == "end":
            global ended
            ended = True
            idle = True
        else:
            operation = change[0]
            client = change[1:]

            if operation == "-":
                try:
                    clients.remove(client)
                except KeyError:
                    print('Client IP address not present in list, cannot remove: ' + client)
            else:
                clients.add(client)

        change = ""

        if (len(clients) <= 0):
            idle = True
        

def signal_handler(sig, frame):
    print('Interrupt received.')
    global ended
    ended = True

    global clients
    clients.clear()

signal.signal(signal.SIGINT, signal_handler)

if __name__ == "__main__":
    args = get_args()
    main(args)