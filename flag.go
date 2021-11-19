package main

import (
	"flag"
	"fmt"
	"strings"
)

const (
	FLAG_SOURCE         = "src"
	FLAG_DESTINATION    = "dst"
	FLAG_REMOVE_TAGS    = "rmtag"
	FLAG_COPY_TAGS      = "cptag"
	FLAG_COPY_TITLE     = "title"
	FLAG_COPY_ALIASES   = "alias"
	FLAG_CONVERT_LINKS  = "link"
	FLAG_REMOVE_COMMENT = "cmmt"
	FLAG_PUBLISHABLE    = "pub"
	FLAG_REMOVE_H1      = "rmh1"
	FLAG_STRICT_REF     = "strictref"
	FLAG_OBSIDIAN_USAGE = "obs"
	FLAG_STANDARD_USAGE = "std"
	FLAG_VERSION        = "version"
	FLAG_DEBUG          = "debug"
)

type flagBundle struct {
	src         string
	dst         string
	rmtag       bool
	cptag       bool
	title       bool
	alias       bool
	link        bool
	cmmt        bool
	publishable bool
	rmH1        bool
	strictref   bool
	obs         bool
	std         bool
	ver         bool
	debug       bool
}

type mainErrKind int

const (
	MAIN_ERR_UNEXPECTED = iota + 1
	MAIN_ERR_KIND_SOURCE_NOT_SET
	MAIN_ERR_KIND_DESTINATION_NOT_SET
	MAIN_ERR_KIND_STRICTREF_NEEDS_LINK
	MAIN_ERR_KIND_INVALID_SOURCE_FORMAT
	MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT
)

type mainErr interface {
	error
	Kind() mainErrKind
}

type mainErrImpl struct {
	kind    mainErrKind
	message string
}

func (e *mainErrImpl) Error() string {
	return e.message
}

func (e *mainErrImpl) Kind() mainErrKind {
	return e.kind
}

func newMainErr(kind mainErrKind) mainErr {
	err := new(mainErrImpl)
	switch kind {
	case MAIN_ERR_KIND_SOURCE_NOT_SET:
		err.message = fmt.Sprintf("%s was not set", FLAG_SOURCE)
	case MAIN_ERR_KIND_DESTINATION_NOT_SET:
		err.message = fmt.Sprintf("%s was not set", FLAG_DESTINATION)
	case MAIN_ERR_KIND_STRICTREF_NEEDS_LINK:
		err.message = fmt.Sprintf("%s set but not %s", FLAG_STRICT_REF, FLAG_CONVERT_LINKS)
	case MAIN_ERR_KIND_INVALID_SOURCE_FORMAT:
		err.message = fmt.Sprintf("%s shouldn't begin with \"-\"", FLAG_SOURCE)
	case MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT:
		err.message = fmt.Sprintf("%s shouldn't begin with \"-\"", FLAG_DESTINATION)
	default:
		err.kind = MAIN_ERR_UNEXPECTED
		err.message = "unexpected error"
		return err
	}
	err.kind = kind
	return err
}

func initFlags(flagset *flag.FlagSet, flags *flagBundle) {
	flagset.StringVar(&flags.src, FLAG_SOURCE, ".", "source directory")
	flagset.StringVar(&flags.dst, FLAG_DESTINATION, ".", "destination directory")
	flagset.BoolVar(&flags.rmtag, FLAG_REMOVE_TAGS, false, "remove tag")
	flagset.BoolVar(&flags.cptag, FLAG_COPY_TAGS, false, "copy tag to tags field of front matter")
	flagset.BoolVar(&flags.title, FLAG_COPY_TITLE, false, "copy h1 content to title field of front matter")
	flagset.BoolVar(&flags.alias, FLAG_COPY_ALIASES, false, "copy add h1 content to aliases field of front matter")
	flagset.BoolVar(&flags.link, FLAG_CONVERT_LINKS, false, "convert obsidian internal and external links to external links in the usual format")
	flagset.BoolVar(&flags.cmmt, FLAG_REMOVE_COMMENT, false, "remove obsidian comment")
	flagset.BoolVar(&flags.publishable, FLAG_PUBLISHABLE, false, "publish: true -> draft: false, publish: false -> draft: true, no publish field -> draft: true. If draft explicitly specified, then leave it as is.")
	flagset.BoolVar(&flags.rmH1, FLAG_REMOVE_H1, false, "remove H1")
	flagset.BoolVar(&flags.strictref, FLAG_STRICT_REF, false, fmt.Sprintf("return error when ref target is not found. available only when %s is on", FLAG_CONVERT_LINKS))
	flagset.BoolVar(&flags.obs, FLAG_OBSIDIAN_USAGE, false, "alias of -cptag -title -alias")
	flagset.BoolVar(&flags.std, FLAG_STANDARD_USAGE, false, "alias of -cptag -rmtag -title -alias -link -cmmt -pub -strictref")
	flagset.BoolVar(&flags.ver, FLAG_VERSION, false, "display the version currently installed")
	flagset.BoolVar(&flags.debug, FLAG_DEBUG, false, "display error message for developers")
}

// 実行前に↓が必要
// 1. initFlags(flagset, flags)
// 2. flag の値の設定
// 	- flag.CommandLine => flag.Parse()
//	- それ以外 => flagset.Set("フラグ名", "フラグの値を表す文字列")
func setFlags(flagset *flag.FlagSet, flags *flagBundle) {
	orgFlag := *flags
	setflags := make(map[string]struct{})
	flagset.Visit(func(f *flag.Flag) {
		setflags[f.Name] = struct{}{}
	})

	if flags.obs || flags.std {
		flags.cptag = true
		flags.title = true
		flags.alias = true
	}
	if flags.std {
		flags.rmtag = true
		flags.link = true
		flags.cmmt = true
		flags.publishable = true
		flags.strictref = true
	}

	if _, ok := setflags[FLAG_COPY_TAGS]; ok {
		flags.cptag = orgFlag.cptag
	}
	if _, ok := setflags[FLAG_COPY_TITLE]; ok {
		flags.title = orgFlag.title
	}
	if _, ok := setflags[FLAG_COPY_ALIASES]; ok {
		flags.alias = orgFlag.alias
	}
	if _, ok := setflags[FLAG_REMOVE_TAGS]; ok {
		flags.rmtag = orgFlag.rmtag
	}
	if _, ok := setflags[FLAG_CONVERT_LINKS]; ok {
		flags.link = orgFlag.link
	}
	if _, ok := setflags[FLAG_REMOVE_COMMENT]; ok {
		flags.cmmt = orgFlag.cmmt
	}
	if _, ok := setflags[FLAG_PUBLISHABLE]; ok {
		flags.publishable = orgFlag.publishable
	}
	if _, ok := setflags[FLAG_STRICT_REF]; ok {
		flags.strictref = orgFlag.strictref
	}
}

func verifyFlags(flags *flagBundle) error {
	if flags.src == "" {
		return newMainErr(MAIN_ERR_KIND_SOURCE_NOT_SET)
	}
	if flags.dst == "" {
		return newMainErr(MAIN_ERR_KIND_DESTINATION_NOT_SET)
	}
	if strings.HasPrefix(flags.src, "-") {
		return newMainErr(MAIN_ERR_KIND_INVALID_SOURCE_FORMAT)
	}
	if strings.HasPrefix(flags.dst, "-") {
		return newMainErr(MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT)
	}
	if flags.strictref && !flags.link {
		return newMainErr(MAIN_ERR_KIND_STRICTREF_NEEDS_LINK)
	}
	return nil
}
