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
		return fmt.Errorf("file.Stat failed %w", err)
	}

	content := make([]byte, fileinfo.Size())
	_, err = file.Read(content)
	if err != nil {
		return fmt.Errorf("file.Write failed: %w", err)
	}

	yml, body := splitMarkdown([]rune(string(content)))
	frontmatter := new(frontMatter)
	title := ""
	tags := make(map[string]struct{})

	newContent, err := convert(body, vault, &title, tags, *flags)
	if err != nil {
		return fmt.Errorf("convert failed: %w", err)
	}

	if err := yaml.Unmarshal(yml, frontmatter); err != nil {
		return fmt.Errorf("yaml.Unmarshal failed: %w", err)
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
		return fmt.Errorf("yaml.Marshal failed: %w", err)
	}

	newfile, err := os.Create(newpath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", newpath, err)
	}
	defer newfile.Close()
	fmt.Fprintf(newfile, "---\n%s---\n%s", string(yml), string(newContent))
	return nil
}

func convert(raw []rune, vault string, title *string, tags map[string]struct{}, flags flagBundle) (output []rune, err error) {
	output = raw
	if flags.cptag {
		_, err = NewTagFinder(tags).Convert(output)
		if err != nil {
			return nil, fmt.Errorf("TagFinder failed: %w", err)
		}
	}
	if flags.rmtag {
		output, err = NewTagRemover().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("TagRemover failed: %w", err)
		}
	}
	if flags.cmmt {
		output, err = NewCommentEraser().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("CommentEraser failed: %w", err)
		}
	}
	if flags.link {
		output, err = NewLinkConverter(vault).Convert(output)
		if err != nil {
			return nil, fmt.Errorf("LinkConverter failed: %w", err)
		}
	}
	if flags.title {
		titleFoundFrom, _ := NewTagRemover().Convert(output)
		if err != nil {
			return nil, fmt.Errorf("preprocess TagRemover for finding titles failed: %w", err)
		}
		_, err = NewTitleFinder(title).Convert(titleFoundFrom)
		if err != nil {
			return nil, fmt.Errorf("TitleFinder failed: %w", err)
		}
	}
	return output, nil
}
