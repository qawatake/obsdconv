package main

import "flag"

func initFlags(flagset *flag.FlagSet, flags *flagBundle) {
	flagset.StringVar(&flags.src, FLAG_SOURCE, ".", "source directory")
	flagset.StringVar(&flags.dst, FLAG_DESTINATION, ".", "destination directory")
	flagset.BoolVar(&flags.rmtag, FLAG_REMOVE_TAGS, false, "remove tag")
	flagset.BoolVar(&flags.cptag, FLAG_COPY_TAGS, false, "copy tag to tags field of front matter")
	flagset.BoolVar(&flags.title, FLAG_COPY_TITLE, false, "copy h1 content to title field of front matter")
	flagset.BoolVar(&flags.alias, FLAG_COPY_ALIASES, false, "copy add h1 content to aliases field of front matter")
	flagset.BoolVar(&flags.link, FLAG_CONVERT_LINKS, false, "convert obsidian internal and external links to external links in the usual format")
	flagset.BoolVar(&flags.cmmt, FLAG_REMOVE_COMMENT, false, "remove obsidian comment")
	flagset.BoolVar(&flags.obs, FLAG_OBSIDIAN_USAGE, false, "alias of -cptag -title -alias")
	flagset.BoolVar(&flags.cmmn, FLAG_COMMON_USAGE, false, "alias of -cptag -rmtag -title -alias -link -cmmt")
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

	if flags.obs || flags.cmmn {
		flags.cptag = true
		flags.title = true
		flags.alias = true
	}
	if flags.cmmn {
		flags.rmtag = true
		flags.link = true
		flags.cmmt = true
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
	return nil
}
