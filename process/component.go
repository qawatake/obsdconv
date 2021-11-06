package process

type BodyConvAuxOut interface{}

type BodyConverter interface {
	ConvertBody(raw []rune) (output []rune, aux BodyConvAuxOut, err error)
}

type YamlConvAuxIn interface{}

type YamlConverter interface {
	ConvertYAML(raw []byte, aux YamlConvAuxIn) (output []byte, err error)
}

type ArgPasser interface {
	PassArg(frombody BodyConvAuxOut) (toyaml YamlConvAuxIn, err error)
}
