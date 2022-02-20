package main

import (
	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
	"github.com/qawatake/obsdconv/process"
)

type bodyConvAuxOutImpl struct {
	title string
	tags  map[string]struct{}
}

func newBodyConvAuxOutImpl(title string, tags map[string]struct{}) *bodyConvAuxOutImpl {
	return &bodyConvAuxOutImpl{
		title: title,
		tags:  tags,
	}
}

type bodyConverterImpl struct {
	db      convert.PathDB
	cptag   bool
	rmtag   bool
	cmmt    bool
	title   bool
	link    bool
	rmH1    bool
	baseUrl string
}

func newBodyConverterImpl(db convert.PathDB, cptag bool, rmtag bool, cmmt bool, title bool, link bool, rmH1 bool, baseUrl string) *bodyConverterImpl {
	c := new(bodyConverterImpl)
	c.db = db
	c.cptag = cptag
	c.rmtag = rmtag
	c.cmmt = cmmt
	c.title = title
	c.link = link
	c.rmH1 = rmH1
	c.baseUrl = baseUrl
	return c
}

func (c *bodyConverterImpl) ConvertBody(raw []rune, selfRelativePath string) (output []rune, aux process.BodyConvAuxOut, err error) {
	output = raw
	title := ""
	tags := make(map[string]struct{})

	if c.cptag {
		_, err = convert.NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, nil, errors.Wrap(err, "TagFinder failed")
		}
	}
	if c.title {
		titleFoundFrom, err := convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, nil, errors.Wrap(err, "preprocess TagRemover for finding titles failed")
		}
		titleFoundFrom, err = convert.NewLinkPlainConverter().Convert(titleFoundFrom)
		if err != nil {
			return nil, nil, errors.Wrap(err, "preprocess InternalLinkPlainConverter for finding titles failed")
		}
		_, err = convert.NewTitleFinder(&title).Convert(titleFoundFrom)
		if err != nil {
			return nil, nil, errors.Wrap(err, "TitleFinder failed")
		}
	}
	if c.rmtag {
		output, err = convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, nil, errors.Wrap(err, "TagRemover failed")
		}
	}
	if c.cmmt {
		output, err = convert.NewCommentEraser().Convert(output)
		if err != nil {
			return nil, nil, errors.Wrap(err, "CommentEraser failed")
		}
	}
	if c.link {
		if c.baseUrl == "" {
			output, err = convert.NewLinkConverter(c.db).Convert(output)
			if err != nil {
				return nil, nil, errors.Wrap(err, "LinkConverter failed")
			}
		} else {
			// fmt.Println(newpath)
			db := convert.WrapForTrimmingSuffixMd(convert.WrapForSettingBaseUrl(c.baseUrl, convert.WrapForUsingSelfForEmptyFileId(selfRelativePath, c.db)))
			output, err = convert.NewLinkConverter(db).Convert(output)
			if err != nil {
				return nil, nil, errors.Wrap(err, "LinkConverter failed")
			}
		}
	}
	if c.rmH1 {
		output, err = convert.NewH1Remover().Convert(output)
		if err != nil {
			return nil, nil, errors.Wrap(err, "H1Remover failed")
		}
	}

	aux = newBodyConvAuxOutImpl(title, tags)
	return output, aux, nil
}
