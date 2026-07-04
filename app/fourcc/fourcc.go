package main

/*
#cgo CFLAGS: -I/ffmpeg/include
#cgo LDFLAGS: -L/ffmpeg/lib -lavcodec -lv4l2

#include <libavcodec/avcodec.h>
#include <linux/videodev2.h>
#include <libv4l2.h>
*/
import "C"
import (
	"encoding/json"
	"log"
	"os"
)

type Mapping struct {
	Format  string
	FourCC  uint32
	CodecID uint32
	PixFmt  int32
}

func main() {
	mapping := make([]Mapping, 0, 15) // TODO: command-line argument to include extra formats

	// formats handled by libv4lconvert are listed below:
	// https://www.kernel.org/doc/html/v7.1/userspace-api/media/v4l/libv4l-introduction.html#libv4lconvert
	// these won't be included below as FFmpeg won't be used for them

	// COMPRESSED VIDEO TYPES

	mapping = append(mapping, Mapping{
		Format:  "HEVC",
		FourCC:  C.V4L2_PIX_FMT_HEVC,
		CodecID: C.AV_CODEC_ID_HEVC,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "H264",
		FourCC:  C.V4L2_PIX_FMT_H264,
		CodecID: C.AV_CODEC_ID_H264,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "H263",
		FourCC:  C.V4L2_PIX_FMT_H263,
		CodecID: C.AV_CODEC_ID_H263,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "VP90",
		FourCC:  C.V4L2_PIX_FMT_VP9,
		CodecID: C.AV_CODEC_ID_VP9,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "VP80",
		FourCC:  C.V4L2_PIX_FMT_VP8,
		CodecID: C.AV_CODEC_ID_VP8,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "MPG4",
		FourCC:  C.V4L2_PIX_FMT_MPEG4,
		CodecID: C.AV_CODEC_ID_MPEG4,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "MPG2",
		FourCC:  C.V4L2_PIX_FMT_MPEG2,
		CodecID: C.AV_CODEC_ID_MPEG2VIDEO,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		Format:  "MPG1",
		FourCC:  C.V4L2_PIX_FMT_MPEG1,
		CodecID: C.AV_CODEC_ID_MPEG1VIDEO,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	// RAW VIDEO TYPES

	mapping = append(mapping, Mapping{
		Format:  "NV12",
		FourCC:  C.V4L2_PIX_FMT_NV12,
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_NV12,
	})

	if (len(os.Args) > 1) && (os.Args[1] == "-a") {
		mapping = append(mapping, Mapping{
			Format:  "MJPG",
			FourCC:  C.V4L2_PIX_FMT_MJPEG,
			CodecID: C.AV_CODEC_ID_MJPEG,
			PixFmt:  C.AV_PIX_FMT_NONE,
		})

		mapping = append(mapping, Mapping{
			Format:  "YUYV",
			FourCC:  C.V4L2_PIX_FMT_YUYV,
			CodecID: C.AV_CODEC_ID_RAWVIDEO,
			PixFmt:  C.AV_PIX_FMT_YUYV422,
		})

		mapping = append(mapping, Mapping{
			Format:  "RGB3",
			FourCC:  C.V4L2_PIX_FMT_RGB24,
			CodecID: C.AV_CODEC_ID_RAWVIDEO,
			PixFmt:  C.AV_PIX_FMT_RGB24,
		})

		mapping = append(mapping, Mapping{
			Format:  "BGR3",
			FourCC:  C.V4L2_PIX_FMT_BGR24,
			CodecID: C.AV_CODEC_ID_RAWVIDEO,
			PixFmt:  C.AV_PIX_FMT_BGR24,
		})

		mapping = append(mapping, Mapping{
			Format:  "YU12",
			FourCC:  C.V4L2_PIX_FMT_YUV420,
			CodecID: C.AV_CODEC_ID_RAWVIDEO,
			PixFmt:  C.AV_PIX_FMT_YUV420P,
		})
	}

	file, err := os.Create("fourcc.json")
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(mapping)
	if err != nil {
		log.Fatal(err)
	}
}
