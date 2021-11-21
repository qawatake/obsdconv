package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/qawatake/obsdconv/process"
)

type argPasserFunc func(frombody process.BodyConvAuxOut) (toyaml process.YamlConvAuxIn, err error)

func (passer argPasserFunc) PassArg(frombody process.BodyConvAuxOut) (toyaml process.YamlConvAuxIn, err error) {
	return passer(frombody)
}

func passArg(frombody process.BodyConvAuxOut) (toyaml process.YamlConvAuxIn, err error) {
	title := ""
	alias := ""
	var newtags []string
	if v, ok := frombody.(*bodyConvAuxOutImpl); !ok {
		return nil, errors.New("frombody (process.BodyConvAuxOutImpl) cannot converted to process.YamlConvAuxInImpl")
	} else {
		title = v.title
		alias = v.title
		newtags = make([]string, 0, len(v.tags))
		for tg := range v.tags {
			newtags = append(newtags, tg)
		}
		// sort tags
		sort.Slice(newtags, func(i, j int) bool {
			return strings.Compare(newtags[i], newtags[j]) <= 0
		})
	}

	return newYamlConvAuxInImpl(title, alias, newtags), nil
}
