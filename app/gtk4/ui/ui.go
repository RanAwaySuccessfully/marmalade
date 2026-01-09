package ui

import (
	_ "embed"
)

//go:embed style.css
var EmbeddedCSS string

//go:embed app.ui
var App string

//go:embed settings_camera.ui
var SettingsCamera string

//go:embed settings_mediapipe.ui
var SettingsMediaPipe string

//go:embed settings_ports.ui
var SettingsPorts string

//go:embed dialog_about.ui
var DialogAbout string

//go:embed dialog_camerainfo.ui
var DialogCameraInfo string

//go:embed dialog_error.ui
var DialogError string
