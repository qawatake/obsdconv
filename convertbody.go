package main

import (
	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
)

type BodyConverter interface {
	ConvertBody(raw []rune) (output []rune, title string, tags map[string]struct{}, err error)
}

type BodyConverterImpl struct {
	flags *flagBundle
	db    convert.PathDB
}

func (c *BodyConverterImpl) ConvertBody(raw []rune) (output []rune, title string, tags map[string]struct{}, err error) {
	output = raw
	title = ""
	tags = make(map[string]struct{})

	if c.flags.cptag {
		_, err = convert.NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "TagFinder failed")
		}
	}
	if c.flags.rmtag {
		output, err = convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "TagRemover failed")
		}
	}
	if c.flags.cmmt {
		output, err = convert.NewCommentEraser().Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "CommentEraser failed")
		}
	}
	if c.flags.title {
		titleFoundFrom, err := convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "preprocess TagRemover for finding titles failed")
		}
		titleFoundFrom, err = convert.NewInternalLinkPlainConverter().Convert(titleFoundFrom)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "preprocess InternalLinkPlainConverter for finding titles failed")
		}
		_, err = convert.NewTitleFinder(&title).Convert(titleFoundFrom)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "TitleFinder failed")
		}
	}
	if c.flags.link {
		output, err = convert.NewLinkConverter(c.db).Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "LinkConverter failed")
		}
	}
	return output, title, tags, nil
}
