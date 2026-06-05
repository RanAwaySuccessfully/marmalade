import { HandSolver } from "./kalidokit/dist/HandSolver/index.js";
import { PoseSolver } from "./kalidokit/dist/PoseSolver/index.js";

process.on("SIGTERM", () => {
    console.log("[KA +TOAST] Terminating");
    process.exit(0);
});

Bun.connect({
    unix: "./kalidokit.sock",
    socket: {
        data(socket, message) {
            const decoder = new TextDecoder();
            const text = decoder.decode(message);
            const payload = JSON.parse(text);

            let result = {
                Type: 0
            };

            try {
                if (payload.hands) {
                    result.Type = 1;
                    payload.hands.forEach(hand => {
                        if (!hand.handedness) {
                            return;
                        }

                        const handedness = hand.handedness[0].category_name;
                        const hand_result = HandSolver.solve(hand.landmarks, handedness);

                        const entries = Object.entries(hand_result).map(([key, value]) => {
                            key = key.replace("Left", "").replace("Right", "");
                            return [key, value];
                        });

                        const formattedResult = Object.fromEntries(entries);
                        
                        switch (handedness) {
                            case "Left":
                                result.LeftHandData = formattedResult;
                                break;
                            case "Right":
                                result.RightHandData = formattedResult;
                                break;
                        }
                    });
                } else {
                    result.Type = 2;
                    result.PoseData = PoseSolver.solve(payload.world_landmarks, payload.landmarks, {
                        runtime: "mediapipe",
                        enableLegs: true,
                    });
                }
            } catch (error) {
                result.Type = 0;
                console.error(error);
            }

            socket.write(JSON.stringify(result));
        }
    }
});