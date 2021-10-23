package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type errKind int

const (
	ErrKindInvalidInternalLinkContent errKind = iota
)

type Err struct {
	kind errKind
}

type ErrInvalidInternalLinkContent interface {
	IsErrInvalidInternalLinkContent() bool
}

func (e *Err) Error() string {
	return "invalid internal link content"
}

func (e *Err) IsErrInvalidInternalLinkContent() bool {
	return e.kind == ErrKindInvalidInternalLinkContent
}

func newErr(kind errKind) *Err {
	return &Err{kind: kind}
}

func pathMatchScore(path string, filename string) int {
	pp := strings.Split(path, "/")
	ff := strings.Split(filename, "/")
	lpp := len(pp)
	lff := len(ff)
	if lpp < lff {
		return -1
	}
	cur := 0
	for ; cur < lff; cur++ {
		if pp[lpp-1-cur] != ff[lff-1-cur] {
			return -1
		}
	}
	return lpp - cur
}

func findPath(root string, fileId string) (path string, err error) {
	var filename string
	if filepath.Ext(fileId) == "" {
		filename = fileId + ".md"
	} else {
		filename = fileId
	}

	bestscore := -1
	bestmatch := ""

	err = filepath.Walk(root, func(pth string, info fs.FileInfo, err error) error {
		if score := pathMatchScore(pth, filename); score < 0 {
			return nil
		} else if bestscore < 0 || score < bestscore || (score == bestscore && strings.Compare(pth, bestmatch) < 0) {
			bestscore = score
			bestmatch = pth
			return nil
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("filepath.Walk failed: %w", err)
	}
	if bestscore < 0 {
		return "", nil
	}
	path, err = filepath.Rel(root, bestmatch)
	if err != nil {
		return "", fmt.Errorf("filepath.Rel failed: %w", err)
	}
	return path, nil
}

func splitDisplayName(fullname string) (identifier string, displayname string) {
	position := strings.Index(fullname, "|")
	if position < 0 {
		return fullname, ""
	} else {
		identifier := strings.Trim(string(fullname[:position]), " \t")
		displayname := strings.TrimLeft(string(fullname[position:]), "|")
		displayname = strings.Trim(displayname, " \t")
		return identifier, displayname
	}
}

func splitFragments(identifier string) (fileId string, fragments []string, err error) {
	strs := strings.Split(identifier, "#")
	if len(strs) == 1 {
		return strs[0], nil, nil
	}
	fileId = strs[0]
	if len(strings.TrimRight(fileId, " \t")) != len(fileId) {
		return "", nil, newErr(ErrKindInvalidInternalLinkContent)
	}

	fragments = make([]string, len(strs)-1)
	for id, f := range strs[1:] {
		f := strings.Trim(f, " \t")
		fragments[id] = f
	}
	return fileId, fragments, nil
}

func buildLinkText(displayName string, fileId string, fragments []string) (linktext string) {
	if displayName != "" {
		return displayName
	}

	linktext = fileId
	for _, f := range fragments {
		linktext += fmt.Sprintf(" > %s", f)
	}
	return linktext
}

func genExternalLink(root string, content string) (link string, err error) {
	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", fmt.Errorf("splitFragments failed: %w", err)
	}
	path, err := findPath(root, fileId)
	if err != nil {
		return "", fmt.Errorf("findPath failed: %w", err)
	}

	linktext := buildLinkText(displayName, fileId, fragments)
	var ref string
	if fragments == nil {
		ref = path
	} else {
		ref = path + "#" + strings.Join(fragments, "#")
	}

	return fmt.Sprintf("[%s](%s)", linktext, ref), nil
}
