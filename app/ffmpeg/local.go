//go:build ffmpeg_git

package main

/*
#cgo CFLAGS: -I${SRCDIR}/ffmpeg/libavcodec -I${SRCDIR}/ffmpeg/libavutil -I${SRCDIR}/ffmpeg/libswscale
#cgo LDFLAGS: ${SRCDIR}/ffmpeg/libavcodec/libavcodec.so ${SRCDIR}/ffmpeg/libavutil/libavutil.so ${SRCDIR}/ffmpeg/libswscale/libswscale.so -lavcodec -lavutil -lswscale
*/
import "C"
