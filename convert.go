package main

import (
	"github.com/pkg/errors"
	"github.com/qawatake/obsd2hugo/convert"
)

func converts(raw []rune, vault string, title *string, tags map[string]struct{}, flags flagBundle) (output []rune, err error) {
	output = raw
	if flags.cptag {
		_, err = convert.NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, errors.Wrap(err, "TagFinder failed")
		}
	}
	if flags.rmtag {
		output, err = convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, errors.Wrap(err, "TagRemover failed")
		}
	}
	if flags.cmmt {
		output, err = convert.NewCommentEraser().Convert(output)
		if err != nil {
			return nil, errors.Wrap(err, "CommentEraser failed")
		}
	}
	if flags.link {
		output, err = convert.NewLinkConverter(vault).Convert(output)
		if err != nil {
			// fmt.Println(err)
			// e := errors.Cause(err)
			// // e := errors.Cause
			// fmt.Println(e)
			// if _, ok := e.(convert.ErrConvert); ok {
			// 	fmt.Println("success")
			// } else {
			// 	fmt.Println("failed")
			// }
			// os.Exit(1)
			// panic("done")

			return nil, errors.Wrap(err, "LinkConverter failed")
		}
	}
	if flags.title {
		titleFoundFrom, _ := convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, errors.Wrap(err, "preprocess TagRemover for finding titles failed")
		}
		_, err = convert.NewTitleFinder(title).Convert(titleFoundFrom)
		if err != nil {
			return nil, errors.Wrap(err, "TitleFinder failed")
		}
	}
	return output, nil
}
