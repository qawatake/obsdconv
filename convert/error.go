package convert

type ErrConvert interface {
	error
	Line() int
	SetLine(line int)
	// Cause だと errors.Cause ですべて展開されてしまうため
	Source() error
}

type errConvertImpl struct {
	line  int
	cause error
}

func (e *errConvertImpl) Line() int {
	return e.line
}

func (e *errConvertImpl) SetLine(line int) {
	e.line = line
}

func (e *errConvertImpl) Source() error {
	return e.cause
}

func (e *errConvertImpl) Error() string {
	// return fmt.Sprintf("line: %d: %v", e.line, e.cause)
	return e.cause.Error()
}

func newErrConvert(cause error) ErrConvert {
	return &errConvertImpl{cause: cause}
}
