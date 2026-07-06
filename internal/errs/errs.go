package errs

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
