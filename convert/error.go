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
	return e.cause.Error()
}

func newErrConvert(cause error) ErrConvert {
	return &errConvertImpl{cause: cause}
}

type ErrKind uint

const (
	ERR_KIND_UNEXPECTED ErrKind = iota
	ERR_KIND_INVALID_INTERNAL_LINK_CONTENT
	ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL
	ERR_KIND_UNEXPECTED_HREF
	ERR_KIND_INVALID_SHORTHAND_OBSIDIAN_URL
)

type errTransformImpl struct {
	kind    ErrKind
	message string
}

type ErrTransform interface {
	Kind() ErrKind
}

func (e *errTransformImpl) Error() string {
	return e.message
}

func (e *errTransformImpl) Kind() ErrKind {
	return e.kind
}

func newErrTransform(kind ErrKind, msg string) *errTransformImpl {
	return &errTransformImpl{kind: kind, message: msg}
}
