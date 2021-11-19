package main

import (
	"errors"
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

var (
	ErrFlagSourceNotSet       = fmt.Errorf("flag %s was not set", FLAG_SOURCE)
	ErrFlagDestinationNotSet  = fmt.Errorf("flag %s was not set", FLAG_DESTINATION)
	ErrFlagStrictRefNeedsLink = fmt.Errorf("%s set but not %s", FLAG_STRICT_REF, FLAG_CONVERT_LINKS)
)

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
func setFlags(flagset *flag.FlagSet, flags *flagBundle) error {
	orgFlag := *flags
	setflags := make(map[string]struct{})
	flagset.Visit(func(f *flag.Flag) {
		setflags[f.Name] = struct{}{}
	})
	if _, ok := setflags[FLAG_SOURCE]; !ok {
		return ErrFlagSourceNotSet
	}
	if _, ok := setflags[FLAG_DESTINATION]; !ok {
		return ErrFlagDestinationNotSet
	}

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

	if flags.strictref && !flags.link {
		return ErrFlagStrictRefNeedsLink
	}

	return nil
}

func verifyFlags(flags *flagBundle) error {
	if strings.HasPrefix(flags.src, "-") {
		return errors.New(`src shouldn't begin with "-"`)
	}
	if strings.HasPrefix(flags.dst, "-") {
		return errors.New(`dst shouldn't begin with "-"`)
	}
	return nil
}
