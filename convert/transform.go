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

func TransformRepeatingTagsFunc() TransformerFunc {
	return MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0
		}
		return advance
	})
}

func TransformInternalLinkFunc(t InternalLinkTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" { // [[ ]] はスキップ
			return advance, nil, nil
		}
		link, err := t.TransformInternalLink(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformInternalLinkFunc")
		}
		return advance, []rune(link), nil
	}
}

func TransformEmnbedsFunc(t EmbedsTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		if content == "" {
			return advance, nil, nil
		}
		link, err := t.TransformEmbeds(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformEmbedsFunc")
		}
		return advance, []rune(link), nil
	}
}

func TransformExternalLinkFunc(t ExternalLinkTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, displayName, ref := scan.ScanExternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}

		externalLink, err := t.TransformExternalLink(displayName, ref)
		if err != nil {
			return 0, nil, errors.Wrap(err, "t.TransformExternalLink failed")
		}
		return advance, []rune(externalLink), nil
	}
}

func TransformInternalLinkToPlain(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
	advance, content := scan.ScanInternalLink(raw, ptr)
	if advance == 0 {
		return 0, nil, nil
	}
	if content == "" { // [[ ]] はスキップ
		return advance, nil, nil
	}

	identifier, displayName := splitDisplayName(content)
	if displayName != "" {
		return advance, []rune(displayName), nil
	}
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return 0, nil, errors.Wrap(err, "splitFragments failed in TransformInternalLinkFunc")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	return advance, []rune(linktext), nil
}

type InternalLinkTransformer interface {
	TransformInternalLink(content string) (externalLink string, err error)
}

type InternalLinkTransformerImpl struct {
	ExternalLinkGenerator
}

func (t *InternalLinkTransformerImpl) TransformInternalLink(content string) (externalLink string, err error) {
	return t.GenExternalLink(content)
}

type EmbedsTransformer interface {
	TransformEmbeds(content string) (embeddedLink string, err error)
}

type EmbedsTransformerImpl struct {
	ExternalLinkGenerator
}

func (t *EmbedsTransformerImpl) TransformEmbeds(content string) (emnbeddedLink string, err error) {
	externalLink, err := t.GenExternalLink(content)
	if err != nil {
		return "", err
	}
	return "!" + externalLink, nil
}

type ExternalLinkTransformer interface {
	TransformExternalLink(displayName, ref string) (externalLink string, err error)
}

type ExternalLinkTransformerImpl struct {
	PathFinder
}

func (t *ExternalLinkTransformerImpl) TransformExternalLink(displayName, ref string) (externalLink string, err error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", newErrTransform(ERR_KIND_UNEXPECTED, fmt.Sprintf("url.Parse failed: %v", err))
	}

	if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return fmt.Sprintf("[%s](%s)", displayName, ref), nil
	} else if u.Scheme == "obsidian" {
		q := u.Query()
		fileId := q.Get("file")
		if fileId == "" {
			return "", newErrTransform(ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL, fmt.Sprintf("no ref file specified in obsidian url: %s", ref))
		}
		path, err := t.FindPath(fileId)
		if err != nil {
			return "", errors.Wrap(err, "FindPath failed")
		}
		return fmt.Sprintf("[%s](%s)", displayName, path), nil
	} else if u.Scheme == "" && u.Host == "" {
		fileId, fragments, err := splitFragments(ref)
		if err != nil {
			return "", errors.Wrap(err, "splitFragments failed")
		}
		path, err := t.FindPath(fileId)
		if err != nil {
			return "", errors.Wrap(err, "findPath failed")
		}
		var newref string
		if fragments == nil {
			newref = path
		} else {
			newref = path + "#" + strings.Join(fragments, "#")
		}
		return fmt.Sprintf("[%s](%s)", displayName, newref), nil
	} else {
		return "", newErrTransform(ERR_KIND_UNEXPECTED_HREF, fmt.Sprintf("unexpected href: %s", ref))
	}
}
