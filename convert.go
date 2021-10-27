package main

import (
	"fmt"

	"github.com/qawatake/obsd2hugo/convert"
)

func converts(raw []rune, vault string, title *string, tags map[string]struct{}, flags flagBundle) (output []rune, err error) {
	output = raw
	if flags.cptag {
		_, err = convert.NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, fmt.Errorf("TagFinder failed: %w", err)
		}
	}
	if flags.rmtag {
		output, err = convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("TagRemover failed: %w", err)
		}
	}
	if flags.cmmt {
		output, err = convert.NewCommentEraser().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("CommentEraser failed: %w", err)
		}
	}
	if flags.link {
		output, err = convert.NewLinkConverter(vault).Convert(output)
		if err != nil {
			return nil, fmt.Errorf("LinkConverter failed: %w", err)
		}
	}
	if flags.title {
		titleFoundFrom, _ := convert.NewTagRemover().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("preprocess TagRemover for finding titles failed: %w", err)
		}
		_, err = convert.NewTitleFinder(title).Convert(titleFoundFrom)
		if err != nil {
			return nil, fmt.Errorf("TitleFinder failed: %w", err)
		}
	}
	return output, nil
}
