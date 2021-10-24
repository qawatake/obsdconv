package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
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

func TransformExternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, displayName, ref := scanExternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil
		}

		u, err := url.Parse(ref)
		if err != nil {
			fmt.Fprintf(os.Stderr, "url.Parse failed in TransformExternalLinkFunc: %v\n", err)
			return 0, nil
		}

		if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
			return advance, raw[ptr : ptr+advance]

		} else if u.Scheme == "obsidian" {
			q := u.Query()
			fileId := q.Get("file")
			if fileId == "" {
				fmt.Fprintf(os.Stderr, "query file does not exits in obsidian url: %s\n", ref)
				return 0, nil
			}
			path, err := findPath(root, fileId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "findPath failed in TransformExternalLinkFunc: %v\n", err)
				return 0, nil
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, path))

		} else if u.Scheme == "" && u.Host == "" {
			fileId, fragments, err := splitFragments(ref)
			if err != nil {
				fmt.Fprintf(os.Stderr, "splitFragments failed in TransformExternalLinkFunc: %v\n", err)
				return 0, nil
			}
			path, err := findPath(root, fileId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "findPath failed in TransformExternalLinkFunc: %v\n", err)
				return 0, nil
			}
			var newref string
			if fragments == nil {
				newref = path
			} else {
				newref = path + "#" + strings.Join(fragments, "#")
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, newref))

		} else {
			fmt.Fprintf(os.Stderr, "unexpected href: %s\n", ref)
			return 0, nil
		}
	}
}
