package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/qawatake/obsdconv/process"
)

type argPasserImpl struct {
	title bool
	alias bool
}

func newArgPasserImpl(title bool, alias bool) *argPasserImpl {
	return &argPasserImpl{
		title: title,
		alias: alias,
	}
}

func (passer *argPasserImpl) PassArg(frombody process.BodyConvAuxOut) (toyaml process.YamlConvAuxIn, err error) {
	title := ""
	alias := ""
	var newtags []string

	// fetch
	args, ok := frombody.(*bodyConvAuxOutImpl)
	if !ok {
		return nil, errors.New("frombody (process.BodyConvAuxOutImpl) cannot converted to process.YamlConvAuxInImpl")
	}
	if passer.title {
		title = args.title
	}
	if passer.alias {
		alias = args.title
	}
	newtags = make([]string, 0, len(args.tags))
	for tg := range args.tags {
		newtags = append(newtags, tg)
	}
	// sort tags
	sort.Slice(newtags, func(i, j int) bool {
		return strings.Compare(newtags[i], newtags[j]) <= 0
	})

	return newYamlConvAuxInImpl(title, alias, newtags), nil
}
