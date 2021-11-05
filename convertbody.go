package main

import (
	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
)

type BodyConverter interface {
	ConvertBody(raw []rune) (output []rune, title string, tags map[string]struct{}, err error)
}

type BodyConverterImpl struct {
	db    convert.PathDB
	cptag bool
	rmtag bool
	cmmt  bool
	title bool
	link  bool
}

func NewBodyConverterImpl(db convert.PathDB, cptag bool, rmtag bool, cmmt bool, title bool, link bool) *BodyConverterImpl {
	c := new(BodyConverterImpl)
	c.db = db
	c.cptag = cptag
	c.rmtag = rmtag
	c.cmmt = cmmt
	c.title = title
	c.link = link
	return c
}

func (c *BodyConverterImpl) ConvertBody(raw []rune) (output []rune, title string, tags map[string]struct{}, err error) {
	output = raw
	title = ""
	tags = make(map[string]struct{})

	if c.cptag {
		_, err = convert.NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "TagFinder failed")
		}
	}
	if c.rmtag {
		output, err = convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "TagRemover failed")
		}
	}
	if c.cmmt {
		output, err = convert.NewCommentEraser().Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "CommentEraser failed")
		}
	}
	if c.title {
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
	if c.link {
		output, err = convert.NewLinkConverter(c.db).Convert(output)
		if err != nil {
			return nil, "", nil, errors.Wrap(err, "LinkConverter failed")
		}
	}
	return output, title, tags, nil
}
