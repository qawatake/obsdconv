package main

type ErrTransform struct {
	path   string
	line   int
	orgErr error
}

func (e *ErrTransform) Path() string {
	return e.path
}

func (e *ErrTransform) Line() int {
	return e.line
}

func (e *ErrTransform) SetPath(path string) {
	e.path = path
}

func (e *ErrTransform) SetLine(line int) {
	e.line = line
}

func (e *ErrTransform) Error() string {
	return e.orgErr.Error()
}

func NewErrTransform(orgErr error) *ErrTransform {
	return &ErrTransform{orgErr: orgErr}
}
