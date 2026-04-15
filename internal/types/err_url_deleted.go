package types

type ErrURLDeleted struct {
	CauseURL ShortURL
	Err      error
}

func (e *ErrURLDeleted) Error() string {
	return "Field with key value: " + string(e.CauseURL) + " already deleted"
}

func (e *ErrURLDeleted) Unwrap() error {
	return e.Err
}
