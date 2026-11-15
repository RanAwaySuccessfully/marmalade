//go:build gstreamer

package webcam

type GStreamer struct{}

func NewGStreamer(fd int) (*GStreamer, error) {
	return nil, nil
}

func (gst *GStreamer) Run(err_ch chan error, detect RunCallback) {}

func (gst *GStreamer) End() {}
