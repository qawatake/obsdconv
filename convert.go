package main

import (
	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
)

type BodyConverter interface {
	ConvertBody(raw []rune) (output *ConvertBodyOutput, err error)
}

type BodyConverterImpl struct {
	flags *flagBundle
	convert.InternalLinkTransformer
	convert.EmbedsTransformer
	convert.ExternalLinkTransformer
}

type ConvertBodyOutput struct {
	text  []rune
	title string
	tags  map[string]struct{}
}

func NewConvertBodyOutput() *ConvertBodyOutput {
	output := new(ConvertBodyOutput)
	output.tags = make(map[string]struct{})
	return output
}

func (c *BodyConverterImpl) ConvertBody(raw []rune) (output *ConvertBodyOutput, err error) {
	output = NewConvertBodyOutput()
	text := raw
	if c.flags.cptag {
		_, err = convert.NewTagFinder(output.tags).Convert(text)
		if err != nil {
			return nil, errors.Wrap(err, "TagFinder failed")
		}
	}
	if c.flags.rmtag {
		text, err = convert.NewTagRemover().Convert(text)
		if err != nil {
			return nil, errors.Wrap(err, "TagRemover failed")
		}
	}
	if c.flags.cmmt {
		text, err = convert.NewCommentEraser().Convert(text)
		if err != nil {
			return nil, errors.Wrap(err, "CommentEraser failed")
		}
	}
	if c.flags.title {
		titleFoundFrom, err := convert.NewTagRemover().Convert(text)
		if err != nil {
			return nil, errors.Wrap(err, "preprocess TagRemover for finding titles failed")
		}
		titleFoundFrom, err = convert.NewInternalLinkPlainConverter().Convert(titleFoundFrom)
		if err != nil {
			return nil, errors.Wrap(err, "preprocess InternalLinkPlainConverter for finding titles failed")
		}
		_, err = convert.NewTitleFinder(&output.title).Convert(titleFoundFrom)
		if err != nil {
			return nil, errors.Wrap(err, "TitleFinder failed")
		}
	}
	if c.flags.link {
		text, err = convert.NewLinkConverter(c.InternalLinkTransformer, c.EmbedsTransformer, c.ExternalLinkTransformer).Convert(text)
		if err != nil {
			return nil, errors.Wrap(err, "LinkConverter failed")
		}
	}
	output.text = text
	return output, nil
}
