package convert

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/unicode/norm"
)

var vaultdict map[string][]string

// findPath で検索するためのデータを設定する
func PrepareVault(vault string) {
	vaultdict = make(map[string][]string)
	filepath.Walk(vault, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		vaultdict[base] = append(vaultdict[base], path)
		return nil
	})
}

func pathMatchScore(path string, filename string) int {
	pp := strings.Split(filepath.ToSlash(path), "/")
	ff := strings.Split(filepath.ToSlash(filename), "/")
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

// 実行前に PrepareVault を呼ぶ必要がある
func findPath(root string, fileId string) (path string, err error) {
	var filename string
	if filepath.Ext(fileId) == "" {
		filename = fileId + ".md"
	} else {
		filename = fileId
	}

	base := filepath.Base(filename)
	paths, ok := vaultdict[base]
	if !ok || len(paths) == 0 {
		return "", nil
	}

	bestscore := -1
	bestmatch := ""
	for _, pth := range paths {
		pth = norm.NFC.String(pth)
		if score := pathMatchScore(pth, filename); score < 0 {
			continue
		} else if bestmatch == "" || score < bestscore || (score == bestscore && strings.Compare(pth, bestmatch) < 0) {
			bestscore = score
			bestmatch = pth
			continue
		}
	}

	if bestscore < 0 {
		return "", nil
	}
	path, err = filepath.Rel(root, bestmatch)
	if err != nil {
		return "", newErrTransform(ERR_KIND_UNEXPECTED, fmt.Sprintf("filepath.Rel failed: %v", err))
	}
	return filepath.ToSlash(path), nil
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
		return "", nil, newErrTransform(ERR_KIND_INVALID_INTERNAL_LINK_CONTENT, fmt.Sprintf("invalid internal link content: %q", identifier))
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

	if fileId != "" {
		linktext = fileId
		for _, f := range fragments {
			linktext += fmt.Sprintf(" > %s", f)
		}
		return linktext
	} else {
		return strings.Join(fragments, " > ")
	}
}

func genExternalLink(root string, content string) (link string, err error) {
	identifier, displayName := splitDisplayName(content)
	fileId, fragments, err := splitFragments(identifier)
	if err != nil {
		return "", errors.Wrap(err, "splitFragments failed")
	}
	path, err := findPath(root, fileId)
	if err != nil {
		return "", errors.Wrap(err, "findPath failed")
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
