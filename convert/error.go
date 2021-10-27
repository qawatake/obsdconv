package convert

import "fmt"

type ErrTransform interface {
	error
	Path() string
	Line() int
	SetPath(path string)
	SetLine(line int)
	Cause() error
}

type errTransformImpl struct {
	path   string
	line   int
	orgErr error
}

func (e *errTransformImpl) Path() string {
	return e.path
}

func (e *errTransformImpl) Line() int {
	return e.line
}

func (e *errTransformImpl) SetPath(path string) {
	e.path = path
}

func (e *errTransformImpl) SetLine(line int) {
	e.line = line
}

func (e *errTransformImpl) Cause() error {
	return e.orgErr
}

func (e *errTransformImpl) Error() string {
	return fmt.Sprintf("[ERROR] path: %s, line: %d: %v", e.path, e.line, e.orgErr)
}

func newErrTransform(orgErr error) ErrTransform {
	return &errTransformImpl{orgErr: orgErr}
}
