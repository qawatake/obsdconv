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
