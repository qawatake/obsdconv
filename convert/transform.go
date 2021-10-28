package convert

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/qawatake/obsd2hugo/scan"
)

func currentLine(raw []rune, ptr int) (linenum int) {
	return strings.Count(string(raw[:ptr]), "\n") + 1
}

func TransformNone(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	return 1, raw[ptr : ptr+1], nil
}

func TransformInternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" { // [[ ]] はスキップ
			return advance, nil, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			err := errors.Wrap(err, "genExternalLink failed in TransformInternalLinkFunc")
			return 0, nil, err
		}
		return advance, []rune(link), nil
	}
}

func TransformEmbedsFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" {
			return advance, nil, nil
		}
		link, err := genExternalLink(root, content)
		if err != nil {
			err := errors.Wrap(err, "genExternalLink failed in TransformEmbedsFunc")
			return 0, nil, err
		}
		return advance, []rune("!" + link), nil
	}
}

func TransformExternalLinkFunc(root string) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, displayName, ref := scan.ScanExternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}

		u, err := url.Parse(ref)
		if err != nil {
			err := errors.Wrap(err, "url.Parse failed in TransformExternalLinkFunc")
			return 0, nil, err
		}

		if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
			return advance, raw[ptr : ptr+advance], nil

		} else if u.Scheme == "obsidian" {
			q := u.Query()
			fileId := q.Get("file")
			if fileId == "" {
				err := errors.Errorf("query file does not exits in obsidian url: %s", ref)
				return 0, nil, err
			}
			path, err := findPath(root, fileId)
			if err != nil {
				err := errors.Wrap(err, "findPath failed in TransformExternalLinkFunc")
				return 0, nil, err
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, path)), nil

		} else if u.Scheme == "" && u.Host == "" {
			fileId, fragments, err := splitFragments(ref)
			if err != nil {
				err := errors.Wrap(err, "splitFragments failed in TransformExternalLinkFunc")
				return 0, nil, err
			}
			path, err := findPath(root, fileId)
			if err != nil {
				err := errors.Wrap(err, "findPath failed in TransformExternalLinkFunc")
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
			err := errors.Errorf("unexpected href: %s", ref)
			return 0, nil, err
		}
	}
}
