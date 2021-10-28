package convert

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/scan"
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
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformInternalLinkFunc")
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
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformEmbedsFunc")
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
			return 0, nil, newErrTransform(ERR_KIND_UNEXPECTED, fmt.Sprintf("url.Parse failed in TransformExternalLinkFunc: %v", err))
		}

		if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
			return advance, raw[ptr : ptr+advance], nil

		} else if u.Scheme == "obsidian" {
			q := u.Query()
			fileId := q.Get("file")
			if fileId == "" {
				return 0, nil, newErrTransform(ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL, fmt.Sprintf("no ref file specified in obsidian url: %s", ref))
			}
			path, err := findPath(root, fileId)
			if err != nil {
				return 0, nil, errors.Wrap(err, "findPath failed in TransformExternalLinkFunc")
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, path)), nil

		} else if u.Scheme == "" && u.Host == "" {
			fileId, fragments, err := splitFragments(ref)
			if err != nil {
				return 0, nil, errors.Wrap(err, "splitFragments failed in TransformExternalLinkFunc")
			}
			path, err := findPath(root, fileId)
			if err != nil {
				return 0, nil, errors.Wrap(err, "findPath failed in TransformExternalLinkFunc")
			}
			var newref string
			if fragments == nil {
				newref = path
			} else {
				newref = path + "#" + strings.Join(fragments, "#")
			}
			return advance, []rune(fmt.Sprintf("[%s](%s)", displayName, newref)), nil

		} else {
			return 0, nil, newErrTransform(ERR_KIND_UNEXPECTED_HREF, fmt.Sprintf("unexpected href: %s", ref))
		}
	}
}
