package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

func splitFragment(identifier string) (fileId string, fragment string) {
	position := strings.Index(identifier, "#")
	if position < 0 {
		return identifier, ""
	} else {
		fileId := strings.Trim(string(identifier[:position]), " \t")
		fragment := strings.TrimLeft(string(identifier[position:]), "#")
		fragment = strings.Trim(fragment, " \t")
		return fileId, fragment
	}
}

func genHugoLink(content string) (link string) {
	identifier, displayName := splitDisplayName(content)
	fileId, fragment := splitFragment(identifier)
	path := findPath(fileId)

	if displayName == "" {
		if fragment == "" {
			displayName = fileId
		} else {
			displayName = fmt.Sprintf("%s > %s", fileId, fragment)
		}
	}

	var ref string
	if fragment != "" {
		ref = fmt.Sprintf("%s#%s", path, fragment)
	} else {
		ref = path
	}

	return fmt.Sprintf("[%s]({{< ref \"%s\" >}})", displayName, ref)
}
