package convert

import "fmt"

type ErrConvert interface {
	error
	Path() string
	Line() int
	SetPath(path string)
	SetLine(line int)
	Cause() error
}

type errConvertImpl struct {
	path   string
	line   int
	orgErr error
}

func (e *errConvertImpl) Path() string {
	return e.path
}

func (e *errConvertImpl) Line() int {
	return e.line
}

func (e *errConvertImpl) SetPath(path string) {
	e.path = path
}

func (e *errConvertImpl) SetLine(line int) {
	e.line = line
}

func (e *errConvertImpl) Cause() error {
	return e.orgErr
}

func (e *errConvertImpl) Error() string {
	return fmt.Sprintf("[ERROR] path: %s, line: %d: %v", e.path, e.line, e.orgErr)
}

func newErrConvert(orgErr error) ErrConvert {
	return &errConvertImpl{orgErr: orgErr}
}
