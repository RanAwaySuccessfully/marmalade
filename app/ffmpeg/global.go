// go:build !ffmpeg-local

package main

/*
#cgo pkg-config: libavcodec libavutil libswscale
#cgo LDFLAGS: -lavcodec -lavutil -lswscale
*/
import "C"
