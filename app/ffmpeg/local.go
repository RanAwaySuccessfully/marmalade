//go:build ffmpeg_git

package main

/*
#cgo CFLAGS: -I${SRCDIR}/ffmpeg
#cgo LDFLAGS: -L${SRCDIR}/ffmpeg/libavcodec -L${SRCDIR}/ffmpeg/libavutil -L${SRCDIR}/ffmpeg/libswscale -lavcodec -lavutil -lswscale
*/
import "C"
