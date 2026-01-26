## VTS 3rd Party API

### VBridger

Do not select the "MediaPipe" option, instead, select "VTube Studio" and type in the relevant IP address. Even though it says "Connect to iPhone", clicking on that button will connect to Marmalade instead.

### VNyan¹

During the first time setup, select Skip when it asks you which options will be used for tracking.

Go to "Settings", then on the "Tracking" tab, select "Phone / ARKit". Choose "VTube Studio" and type in the relevant IP address.

### VSeeFace¹

Select any option during the first time setup (this will be overriden later).

Go to "General Settings" then scroll down to "iFacialMocap/FaceMocap3D/VTube Studio" and set the tracking app to "VTube Studio", then type in the relevant IP address.

### Notes

I haven't tested with other programs yet, but in case it doesn't work or works weirdly, feel free to open an issue and/or feature request.

¹ During testing, these programs worked best when running using Proton 10. You should also install the font arial.ttf by copying it from a Windows installation to the folder `[wine prefix]/drive_c/windows/Fonts/`. Your wine prefix folder will vary (the default one is at `~/.wine`).

## VTS Plugin

Connecting directly to VTube Studio is possible, but Marmalade will (mostly) export the MediaPipe data as-is, with minimal mapping to the Live2D parameters that VTube Studio expects. It's likely you'll need to remap some parameters **and** their sensitivity.

As long as you have this setting enabled, Marmalade will try to connect to VTube Studio (asking for permission if it's the first time). You can manually select if you want to send face tracking and/or hand tracking data.

You might also want to consider OpenSeeFace instead, for this scenario. Check out [VTS's Linux Guide](https://github.com/DenchiSoft/VTubeStudio/wiki/Running-VTS-on-Linux) or use [Facetracker](https://codeberg.org/ZRayEntertainment/Facetracker). Note that even if using Facetracker, you need to check the linked guide in order to setup the `ip.txt` file.

## VMC Protocol

As long as you have this setting enabled, Marmalade will try to send data to any application listening on the configured port. The default port is `39540` (Assistant) but some programs may expect port `39539` (Performer) instead.

You can manually select if you want to send face tracking and/or hand tracking data.

### Warudo

Switch the port on Marmalade > Connection/Port Settings > VMC Protocol to `39539`, then save. That should be all you need.