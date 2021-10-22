package main

import (
	"fmt"
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

func findPath(name string) string {
	var filename string
	switch filepath.Ext(name) {
	case "":
		filename = name + ".md"
	case ".md":
		filename = name
	}
	return filename
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

func genHugoLink(content string) (link string, err error) {
	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", fmt.Errorf("splitFragments failed: %w", err)
	}
	path := findPath(fileId)

	if displayName == "" {
		if fragments == nil {
			displayName = fileId
		} else {
			displayName = fmt.Sprintf("%s > %s", fileId, fragments[0])
		}
	}

	var ref string
	if fragments != nil {
		ref = fmt.Sprintf("%s#%s", path, fragments[0])
	} else {
		ref = path
	}

	return fmt.Sprintf("[%s]({{< ref \"%s\" >}})", displayName, ref), nil
}
