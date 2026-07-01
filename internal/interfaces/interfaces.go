package interfaces

import "io"

type FFMPEG interface {
	Init(codec_id uint32, pix_fmt int32)
	FindHwAccel() []uint32
	UseHwAccel(device string) error
	Ready() error
	Convert(input []byte, output io.Writer) error
	End()
}

type DetailedError struct {
	Where string
	Err   error
}

func (err *DetailedError) Error() string {
	err_str := err.Err.Error()
	return "[MP +TOAST]: Error " + err.Where + ", " + err_str + "\n"
}

func CreateError(where string, err error) error {
	return &DetailedError{
		Where: where,
		Err:   err,
	}
}
