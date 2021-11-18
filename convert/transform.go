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

func TransformInternalLinkFunc(t InternalLinkTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanInternalLink(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		link, err := t.TransformInternalLink(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformInternalLinkFunc")
		}
		return advance, []rune(link), nil
	}
}

func defaultTransformInternalLinkFunc(db PathDB) TransformerFunc {
	return TransformInternalLinkFunc(newInternalLinkTransformerImpl(db))
}

func TransformEmnbedsFunc(t EmbedsTransformer) TransformerFunc {
	return func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, content := scan.ScanEmbeds(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		link, err := t.TransformEmbeds(content)
		if err != nil {
			return 0, nil, errors.Wrap(err, "genExternalLink failed in TransformEmbedsFunc")
		}
		return advance, []rune(link), nil
	}
}

func defaultTransformEmbedsFunc(db PathDB) TransformerFunc {
	return TransformEmnbedsFunc(newEmbedsTransformerImpl(db))
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

func defaultTransformExternalLinkFunc(db PathDB) TransformerFunc {
	return TransformExternalLinkFunc(newExternalLinkTransformerImpl(db))
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
	PathDB
}

func newInternalLinkTransformerImpl(db PathDB) *InternalLinkTransformerImpl {
	return &InternalLinkTransformerImpl{
		PathDB: db,
	}
}

func (t *InternalLinkTransformerImpl) TransformInternalLink(content string) (externalLink string, err error) {
	if content == "" {
		return "", nil // [[ ]] はスキップ
	}

	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", errors.Wrap(err, "splitFragments failed")
	}
	path, err := t.Get(fileId)
	if err != nil {
		return "", errors.Wrap(err, "PathDB.Get failed")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	var ref string
	if fragments == nil {
		ref = path
	} else {
		ref = path + "#" + fragments[len(fragments)-1]
	}

	return fmt.Sprintf("[%s](%s)", linktext, ref), nil
}

type EmbedsTransformer interface {
	TransformEmbeds(content string) (embeddedLink string, err error)
}

type EmbedsTransformerImpl struct {
	PathDB
}

func newEmbedsTransformerImpl(db PathDB) *EmbedsTransformerImpl {
	return &EmbedsTransformerImpl{
		PathDB: db,
	}
}

func (t *EmbedsTransformerImpl) TransformEmbeds(content string) (emnbeddedLink string, err error) {
	if content == "" {
		return "", nil // [[ ]] はスキップ
	}

	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", errors.Wrap(err, "splitFragments failed")
	}
	path, err := t.Get(fileId)
	if err != nil {
		return "", errors.Wrap(err, "PathDB.Get failed")
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	var ref string
	if fragments == nil {
		ref = path
	} else {
		ref = path + "#" + fragments[len(fragments)-1]
	}

	return fmt.Sprintf("![%s](%s)", linktext, ref), nil
}

type ExternalLinkTransformer interface {
	TransformExternalLink(displayName, ref string) (externalLink string, err error)
}

type ExternalLinkTransformerImpl struct {
	PathDB
}

func newExternalLinkTransformerImpl(db PathDB) *ExternalLinkTransformerImpl {
	return &ExternalLinkTransformerImpl{
		PathDB: db,
	}
}

func (t *ExternalLinkTransformerImpl) TransformExternalLink(displayName, ref string) (externalLink string, err error) {
	u, err := url.Parse(ref)
	if err != nil {
		return "", newErrTransform(ERR_KIND_UNEXPECTED, fmt.Sprintf("url.Parse failed: %v", err))
	}

	// ref = 通常のリンク
	if (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		return fmt.Sprintf("[%s](%s)", displayName, ref), nil
	}

	// ref = obsidian URI
	if u.Scheme == "obsidian" {
		q := u.Query()
		fileId := q.Get("file")
		if fileId == "" {
			return "", newErrTransform(ERR_KIND_NO_REF_SPECIFIED_IN_OBSIDIAN_URL, fmt.Sprintf("no ref file specified in obsidian url: %s", ref))
		}
		path, err := t.Get(fileId)
		if err != nil {
			return "", errors.Wrap(err, "PathDB.Get failed")
		}
		return fmt.Sprintf("[%s](%s)", displayName, path), nil
	}

	// ref = fileId
	if u.Scheme == "" && u.Host == "" {
		fileId, fragments, err := splitFragments(ref)
		if err != nil {
			return "", errors.Wrap(err, "splitFragments failed")
		}
		path, err := t.Get(fileId)
		if err != nil {
			return "", errors.Wrap(err, "PathDB.Get failed")
		}
		var newref string
		if fragments == nil {
			newref = path
		} else {
			newref = path + "#" + strings.Join(fragments, "#")
		}
		return fmt.Sprintf("[%s](%s)", displayName, newref), nil
	}

	return "", newErrTransform(ERR_KIND_UNEXPECTED_HREF, fmt.Sprintf("unexpected href: %s", ref))
}
