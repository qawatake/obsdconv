package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	RuneDollar = 0x24 // "$"
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
	FLAG_OBSIDIAN_USAGE = "obs"
	FLAG_COMMON_USAGE   = "cmmn"
)

type flagBundle struct {
	src   string
	dst   string
	rmtag bool
	cptag bool
	title bool
	alias bool
	link  bool
	cmmt  bool
	obs   bool
	cmmn  bool
}

var (
	ErrFlagSourceNotSet      = fmt.Errorf("flag %s was not set", FLAG_SOURCE)
	ErrFlagDestinationNotSet = fmt.Errorf("flag %s was not set", FLAG_DESTINATION)
)

var flags flagBundle

func init() {
	initFlags(flag.CommandLine, &flags)
}

func main() {
	flag.Parse()
	if err := setFlags(flag.CommandLine, &flags); err != nil {
		log.Fatal(err)
	}
	if err := walk(&flags); err != nil {
		log.Fatal(err)
	}
}

func process(vault string, newpath string, flags *flagBundle, file *os.File) error {
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatalf("file.Stat failed %v", err)
	}

	content := make([]byte, fileinfo.Size())
	_, err = file.Read(content)
	if err != nil {
		log.Fatalf("file.Write failed: %v", err)
	}

	yml, body := splitMarkdown([]rune(string(content)))
	frontmatter := new(frontMatter)
	title := ""
	tags := make(map[string]struct{})

	newContent := convert(body, vault, &title, tags, *flags)

	if err := yaml.Unmarshal(yml, frontmatter); err != nil {
		log.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if flags.title {
		frontmatter.Title = title
	}
	if flags.alias {
		frontmatter.Aliases = append(frontmatter.Aliases, frontmatter.Title)
	}
	if flags.cptag {
		for key := range tags {
			frontmatter.Tags = append(frontmatter.Tags, key)
		}
	}
	yml, err = yaml.Marshal(frontmatter)
	if err != nil {
		log.Fatalf("yaml.Marshal failed: %v", err)
	}

	newfile, err := os.Create(newpath)
	if err != nil {
		return err
	}
	defer newfile.Close()
	fmt.Fprintf(newfile, "---\n%s---\n%s", string(yml), string(newContent))
	return nil
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

func convert(raw []rune, vault string, title *string, tags map[string]struct{}, flags flagBundle) (output []rune) {
	output = raw
	if flags.cptag {
		_ = NewTagFinder(tags).Convert(output)
	}
	if flags.rmtag {
		output = NewTagRemover().Convert(output)
	}
	if flags.cmmt {
		output = NewCommentEraser().Convert(output)
	}
	if flags.link {
		output = NewLinkConverter(vault).Convert(output)
	}
	if flags.title {
		titleFoundFrom := NewTagRemover().Convert(output)
		_ = NewTitleFinder(title).Convert(titleFoundFrom)
	}
	return output
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
	flagset.BoolVar(&flags.obs, FLAG_OBSIDIAN_USAGE, false, "alias of -cptag -title -alias")
	flagset.BoolVar(&flags.cmmn, FLAG_COMMON_USAGE, false, "alias of -cptag -rmtag -title -alias -link -cmmt")
}
