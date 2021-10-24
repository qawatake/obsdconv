package main

import (
	"fmt"
	"os"
)

func TransformNone(raw []rune, ptr int) (advance int, tobewritten []rune) {
	return 1, raw[ptr : ptr+1]
}

func TransformInternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, content := scanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		if content == "" { // [[ ]] はスキップ
			return advance, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "genExternalLink failed in TransformInternalLinkFunc: %v", err)
			return 0, nil
		}
		return advance, []rune(link)
	}
}

func TransformEmbedsFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, content := scanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		if content == "" {
			return advance, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "genExternalLink failed in TransformEmbedsFunc: %v", err)
			return 0, nil
		}
		return advance, []rune("!" + link)
	}
}

func TransformTag(raw []rune, ptr int) (advance int, tobewritten []rune) {
	advance, _ = scanTag(raw, ptr)
	if advance == 0 {
		return 0, nil
	}
	return advance, nil
}
