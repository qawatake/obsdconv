package main

import (
	"fmt"
	"net/url"
	"strings"
)

func TransformNone(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	return 1, raw[ptr : ptr+1], nil
}

func TransformInternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" { // [[ ]] はスキップ
			return advance, nil, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
			// fmt.Fprintf(os.Stderr, "genExternalLink failed in TransformInternalLinkFunc: %v", err)
			err := newErrTransform(fmt.Errorf("genExternalLink failed in TransformInternalLinkFunc: %w", err))
			err.SetLine(currentLine(raw, ptr))
			return 0, nil, err
		}
		return advance, []rune(link), nil
	}
}

func TransformEmbedsFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" {
			return advance, nil, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
			// fmt.Fprintf(os.Stderr, "genExternalLink failed in TransformEmbedsFunc: %v", err)
			err := newErrTransform(fmt.Errorf("genExternalLink failed in TransformEmbedsFunc: %w", err))
			err.SetLine(currentLine(raw, ptr))
			return 0, nil, err
		}
		return advance, []rune("!" + link), nil
	}
}

func TransformTag(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	advance, _ = scanTag(raw, ptr)
	if advance == 0 {
		return 0, nil, nil
	}
	return advance, nil, nil
}

func TransformExternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, displayName, ref := scanExternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}

		u, err := url.Parse(ref)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
			// fmt.Fprintf(os.Stderr, "url.Parse failed in TransformExternalLinkFunc: %v\n", err)
			err := newErrTransform(fmt.Errorf("url.Parse failed in TransformExternalLinkFunc: %w", err))
			err.SetLine(currentLine(raw, ptr))
			return 0, nil, err
		}

		if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
			return advance, raw[ptr : ptr+advance], nil

		} else if u.Scheme == "obsidian" {
			q := u.Query()
			fileId := q.Get("file")
			if fileId == "" {
				// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
				// fmt.Fprintf(os.Stderr, "query file does not exits in obsidian url: %s\n", ref)
				err := newErrTransform(fmt.Errorf("query file does not exits in obsidian url: %s", ref))
				err.SetLine(currentLine(raw, ptr))
				return 0, nil, err
			}
			path, err := findPath(root, fileId)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
				// fmt.Fprintf(os.Stderr, "findPath failed in TransformExternalLinkFunc: %v\n", err)
				err := newErrTransform(fmt.Errorf("findPath failed in TransformExternalLinkFunc: %w", err))
				err.SetLine(currentLine(raw, ptr))
				return 0, nil, err
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, path)), nil

		} else if u.Scheme == "" && u.Host == "" {
			fileId, fragments, err := splitFragments(ref)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
				// fmt.Fprintf(os.Stderr, "splitFragments failed in TransformExternalLinkFunc: %v\n", err)
				err := newErrTransform(fmt.Errorf("splitFragments failed in TransformExternalLinkFunc: %w", err))
				err.SetLine(currentLine(raw, ptr))
				return 0, nil, err
			}
			path, err := findPath(root, fileId)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
				// fmt.Fprintf(os.Stderr, "findPath failed in TransformExternalLinkFunc: %v\n", err)
				err := newErrTransform(fmt.Errorf("findPath failed in TransformExternalLinkFunc: %w", err))
				err.SetLine(currentLine(raw, ptr))
				return 0, nil, err
			}
			var newref string
			if fragments == nil {
				newref = path
			} else {
				newref = path + "#" + strings.Join(fragments, "#")
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, newref)), nil

		} else {
			// fmt.Fprintf(os.Stderr, "line %d: ", currentLine(raw, ptr))
			// fmt.Fprintf(os.Stderr, "unexpected href: %s\n", ref)
			err := newErrTransform(fmt.Errorf("unexpected href: %s", ref))
			err.SetLine(currentLine(raw, ptr))
			return 0, nil, err
		}
	}
}

func TransformComment(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	advance = scanComment(raw, ptr)
	if advance == 0 {
		return 0, nil, nil
	}
	return advance, nil, nil
}
