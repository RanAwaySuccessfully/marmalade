from scipy.spatial.transform import Rotation

def get_eye_left_x(blendshapes):
    eye_left = blendshapes["eyeLookOutLeft"]
    eye_right = blendshapes["eyeLookInLeft"]
    return 0 - eye_left + eye_right

def get_eye_left_y(blendshapes):
    eye_up = blendshapes["eyeLookUpLeft"]
    eye_down = blendshapes["eyeLookDownLeft"]
    return 0 - eye_up + eye_down

def get_eye_right_x(blendshapes):
    eye_right = blendshapes["eyeLookOutRight"]
    eye_left = blendshapes["eyeLookInRight"]
    return 0 - eye_left + eye_right

def get_eye_right_y(blendshapes):
    eye_up = blendshapes["eyeLookUpRight"]
    eye_down = blendshapes["eyeLookDownRight"]
    return 0 - eye_up + eye_down


def create_blendshapes_dict(blendshape_list):
    shapes = {}
    for shape in blendshape_list:
        shapes[shape.category_name] = shape.score
    return shapes


def compute_params_from_blendshapes(blendshape_list):
    # Note left/right switched between mediapipe and vtube studio parameters
    blendshapes = create_blendshapes_dict(blendshape_list)

    return {
        "EyeLeft": {
            "x": get_eye_right_x(blendshapes),
            "y": get_eye_right_y(blendshapes),
            "z": 0,
        },
        "EyeRight": {
            "x": get_eye_left_x(blendshapes),
            "y": get_eye_left_y(blendshapes),
            "z": 0,
        }
    }


def compute_params_from_matrix(isometry):
    # Face Position
    translation_vector = isometry[:3, 3]

    # Face Angle
    # Compute rotation from transform isometry matrix
    rotation_matrix = isometry[:3, :3]
    r = Rotation.from_matrix(rotation_matrix)
    angles = r.as_euler("zyx", degrees=True)

    return {
        "Position": {
            "x": translation_vector[0],
            "y": translation_vector[1],
            "z": -translation_vector[2],
        },
        "Rotation": {
            "x": -angles[1],
            "y": angles[2],
            "z": -angles[0],
        }
    }
