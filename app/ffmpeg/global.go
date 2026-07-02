//go:build !ffmpeg_git

package main

/*
#cgo pkg-config: libavcodec libavutil libswscale
#cgo LDFLAGS: -lavcodec -lavutil -lswscale
*/
import "C"
