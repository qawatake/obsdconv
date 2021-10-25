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

var srcflag string
var dstflag string
var rmtagflag bool
var cptagflag bool
var titleflag bool
var aliasflag bool
var linkflag bool
var cmmtflag bool
var obsflag bool
var cmmnflag bool

func init() {
	flag.StringVar(&srcflag, "src", ".", "source directory")
	flag.StringVar(&dstflag, "dst", ".", "destination directory")
	flag.BoolVar(&rmtagflag, "rmtag", false, "remove tag")
	flag.BoolVar(&cptagflag, "cptag", false, "copy tag to tags field of front matter")
	flag.BoolVar(&titleflag, "title", false, "copy h1 content to title field of front matter")
	flag.BoolVar(&aliasflag, "alias", false, "copy add h1 content to aliases field of front matter")
	flag.BoolVar(&linkflag, "link", false, "convert obsidian internal and external links to external links in the usual format")
	flag.BoolVar(&cmmtflag, "cmmt", false, "remove obsidian comment")
	flag.BoolVar(&obsflag, "obs", false, "alias of -cptag -title -alias")
	flag.BoolVar(&cmmnflag, "cmmn", false, "alias of -cptag -rmtag -title -alias -link -cmmt")
}

func main() {
	flag.Parse()
	setFlags()

	if len(flag.Args()) != 1 {
		log.Fatal("引数の個数が不正です")
	}
	filename := flag.Args()[0]
	root := "."

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("os.Create failed %v", err)
	}
	defer file.Close()
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
	tags := make(map[string]struct{})

	c := NewConverter(root, &(frontmatter.Title), tags)
	newContent := c.Convert(body)

	if err := yaml.Unmarshal(yml, frontmatter); err != nil {
		log.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	frontmatter.Aliases = append(frontmatter.Aliases, frontmatter.Title)
	for key := range tags {
		frontmatter.Tags = append(frontmatter.Tags, key)
	}
	yml, err = yaml.Marshal(frontmatter)
	if err != nil {
		log.Fatalf("yaml.Marshal failed: %v", err)
	}
	fmt.Printf("Front Matter: <<\n%v>>\n", string(yml))

	newFile, err := os.Create("new." + filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer newFile.Close()
	fmt.Fprintf(newFile, "---\n%s---\n%s", string(yml), string(newContent))
}

func NewConverter(vault string, title *string, tags map[string]struct{}) *Converter {
	*title = ""

	// タグ削除用の Converter を作成
	remover := new(Converter)
	remover.Set(DefaultMiddleware(scanEscaped))
	remover.Set(DefaultMiddleware(scanCodeBlock))
	remover.Set(TransformComment)
	remover.Set(DefaultMiddleware(scanMathBlock))
	remover.Set(TransformExternalLinkFunc(vault))
	remover.Set(TransformInternalLinkFunc(vault))
	remover.Set(TransformEmbedsFunc(vault))
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
	c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(TransformExternalLinkFunc(vault))
	c.Set(TransformInternalLinkFunc(vault))
	c.Set(TransformEmbedsFunc(vault))
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
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
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, tag := scanTag(raw, ptr)
		if advance == 0 {
			return 0, nil
		}
		tags[tag] = struct{}{}
		return advance, nil
	})
	c.Set(TransformNone)
	return c
}

func setFlags() {
	if obsflag || cmmnflag {
		cptagflag = true
		titleflag = true
		aliasflag = true
	}
	if cmmnflag {
		rmtagflag = true
		linkflag = true
		cmmnflag = true
	}
}
