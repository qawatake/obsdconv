package convert

import "fmt"

type ErrConvert interface {
	error
	Line() int
	SetLine(line int)
	Cause() error
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

func (e *errConvertImpl) Cause() error {
	return e.cause
}

func (e *errConvertImpl) Error() string {
	return fmt.Sprintf("line: %d: %v", e.line, e.cause)
}

func newErrConvert(cause error) ErrConvert {
	return &errConvertImpl{cause: cause}
}
