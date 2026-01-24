package main

/*
#cgo CFLAGS: -I/ffmpeg/include
#cgo LDFLAGS: -L/ffmpeg/lib -lavcodec

#include <libavcodec/avcodec.h>
*/
import "C"
import (
	"encoding/json"
	"log"
	"os"
)

type Mapping struct {
	FourCC  string
	CodecID int
	PixFmt  int
}

func main() {
	mapping := make([]Mapping, 0, 15)

	// COMPRESSED VIDEO TYPES

	mapping = append(mapping, Mapping{
		FourCC:  "MJPG", // V4L2_PIX_FMT_MJPEG
		CodecID: C.AV_CODEC_ID_MJPEG,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "HEVC", // V4L2_PIX_FMT_HEVC
		CodecID: C.AV_CODEC_ID_HEVC,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "H264", // V4L2_PIX_FMT_H264
		CodecID: C.AV_CODEC_ID_H264,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "H263", // V4L2_PIX_FMT_H263
		CodecID: C.AV_CODEC_ID_H263,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "VP90", // V4L2_PIX_FMT_VP9
		CodecID: C.AV_CODEC_ID_VP9,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "VP80", // V4L2_PIX_FMT_VP8
		CodecID: C.AV_CODEC_ID_VP8,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "MPG4", // V4L2_PIX_FMT_MPEG4
		CodecID: C.AV_CODEC_ID_MPEG4,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "MPG2", // V4L2_PIX_FMT_MPEG2
		CodecID: C.AV_CODEC_ID_MPEG2VIDEO,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "MPG1", // V4L2_PIX_FMT_MPEG1
		CodecID: C.AV_CODEC_ID_MPEG1VIDEO,
		PixFmt:  C.AV_PIX_FMT_NONE,
	})

	// RAW VIDEO TYPES

	mapping = append(mapping, Mapping{
		FourCC:  "YUYV", // V4L2_PIX_FMT_YUYV
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_YUYV422,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "RGB3", // V4L2_PIX_FMT_RGB24
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_RGB24,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "BGR3", // V4L2_PIX_FMT_BGR24
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_BGR24,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "NV12", // V4L2_PIX_FMT_NV12
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_NV12,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "YU12", // V4L2_PIX_FMT_YUV420
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_YUV420P,
	})

	mapping = append(mapping, Mapping{
		FourCC:  "YV12", // V4L2_PIX_FMT_YVU420
		CodecID: C.AV_CODEC_ID_RAWVIDEO,
		PixFmt:  C.AV_PIX_FMT_YUV420P,
	})
	// color will be wrong, need to swap U <-> V
	// -vf "shuffleplanes=0:2:1"
	// -vtag YV12

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
