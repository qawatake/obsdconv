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

var flags flagBundle

func init() {
	flag.StringVar(&flags.src, "src", ".", "source directory")
	flag.StringVar(&flags.dst, "dst", ".", "destination directory")
	flag.BoolVar(&flags.rmtag, "rmtag", false, "remove tag")
	flag.BoolVar(&flags.cptag, "cptag", false, "copy tag to tags field of front matter")
	flag.BoolVar(&flags.title, "title", false, "copy h1 content to title field of front matter")
	flag.BoolVar(&flags.alias, "alias", false, "copy add h1 content to aliases field of front matter")
	flag.BoolVar(&flags.link, "link", false, "convert obsidian internal and external links to external links in the usual format")
	flag.BoolVar(&flags.cmmt, "cmmt", false, "remove obsidian comment")
	flag.BoolVar(&flags.obs, "obs", false, "alias of -cptag -title -alias")
	flag.BoolVar(&flags.cmmn, "cmmn", false, "alias of -cptag -rmtag -title -alias -link -cmmt")
}

func main() {
	flag.Parse()
	setFlags()
	if err := walk(&flags); err != nil {
		log.Fatal(err)
	}
}

func convert(vault string, newpath string, flags *flagBundle, file *os.File) error {
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

	c := NewConverter(vault, &title, tags)
	newContent := c.Convert(body)

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

func NewConverter(vault string, title *string, tags map[string]struct{}) *Converter {
	*title = ""

	// タグ削除用の Converter を作成
	remover := new(Converter)
	remover.Set(DefaultMiddleware(scanEscaped))
	remover.Set(DefaultMiddleware(scanCodeBlock))
	if flags.cmmt {
		remover.Set(TransformComment)
	} else {
		remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance = scanComment(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
	}
	remover.Set(DefaultMiddleware(scanMathBlock))
	if flags.link {
		remover.Set(TransformExternalLinkFunc(vault))
		remover.Set(TransformInternalLinkFunc(vault))
		remover.Set(TransformEmbedsFunc(vault))
	} else {
		remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _, _ = scanExternalLink(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
		remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _ = scanInternalLink(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
		remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _ = scanEmbeds(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
	}
	remover.Set(DefaultMiddleware(scanInlineMath))
	remover.Set(DefaultMiddleware(scanInlineCode))
	remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	remover.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, tag := scanTag(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		tags[tag] = struct{}{}
		return advance, nil
	})
	remover.Set(TransformNone)

	// メインの converter を作成
	c := new(Converter)
	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	if flags.cmmt {
		c.Set(TransformComment)
	} else {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance = scanComment(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
	}
	c.Set(DefaultMiddleware(scanMathBlock))
	if flags.link {
		c.Set(TransformExternalLinkFunc(vault))
		c.Set(TransformInternalLinkFunc(vault))
		c.Set(TransformEmbedsFunc(vault))
	} else {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _, _ = scanExternalLink(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _ = scanInternalLink(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, _ = scanEmbeds(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			return advance, raw[ptr : ptr+advance]
		})
	}
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	if flags.rmtag {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, level, _ := scanHeader(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			tagRemoved := remover.Convert(raw[ptr : ptr+advance])
			if level == 1 && *title == "" {
				_, _, *title = scanHeader(tagRemoved, 0)
			}
			return advance, tagRemoved
		})
	} else {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, level, _ := scanHeader(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			tagRemoved := remover.Convert(raw[ptr : ptr+advance])
			if level == 1 && *title == "" {
				_, _, *title = scanHeader(tagRemoved, 0)
			}
			return advance, raw[ptr : ptr+advance]
		})
	}
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	if flags.rmtag {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, tag := scanTag(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			tags[tag] = struct{}{}
			if flags.rmtag {
				return advance, nil
			} else {
				return advance, raw[ptr : ptr+advance]
			}
		})
	} else {
		c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
			advance, tag := scanTag(raw, ptr)
			if advance == 0 {
				return 0, nil
			}
			tags[tag] = struct{}{}
			return advance, raw[ptr : ptr+advance]
		})
	}
	c.Set(TransformNone)
	return c
}

func setFlags() {
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
}
