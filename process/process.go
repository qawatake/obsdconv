package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Processor interface {
	Process(orgpath, newpath string) (err error)
}

type ProcessorImpl struct {
	BodyConverter
	YamlConverter
	ArgPasser
	YamlExaminator
}

func NewProcessor(bc BodyConverter, yc YamlConverter, passer ArgPasser, examinator YamlExaminator) Processor {
	return &ProcessorImpl{
		BodyConverter:  bc,
		YamlConverter:  yc,
		ArgPasser:      passer,
		YamlExaminator: examinator,
	}
}

func (p *ProcessorImpl) Process(orgpath, newpath string) error {

	if filepath.Ext(orgpath) != ".md" {
		file, err := os.Open(orgpath)
		if err != nil {
			return err
		}
		defer file.Close()
		newfile, err := os.Create(newpath)
		if err != nil {
			return err
		}
		defer newfile.Close()
		io.Copy(newfile, file)
		return nil
	}

	readFrom, err := os.Open(orgpath)
	if err != nil {
		return errors.Errorf("failed to open %s", orgpath)
	}
	content, err := io.ReadAll(readFrom)
	if err != nil {
		return errors.New("failed to read file")
	}
	readFrom.Close()

	yml, body := splitMarkdown([]rune(string(content)))

	if ok, err := p.ExamineYaml(yml); err != nil {
		return errors.Wrap(err, "failed to examine yaml front mattter")
	} else if !ok {
		return nil
	}

	output, frombody, err := p.ConvertBody(body)
	if err != nil {
		return errors.Wrap(err, "failed to convert body")
	}

	toyaml, err := p.PassArg(frombody)
	if err != nil {
		return errors.Wrap(err, "failed to pass args from body converter to yaml converter")
	}

	yml, err = p.ConvertYAML(yml, toyaml)
	if err != nil {
		return errors.Wrap(err, "failed to convert yaml")
	}

	// os.Create によってファイルの内容は削除されるので,
	// 変換がすべて正常に行われた後で, 書き込み先のファイルを開く
	writeTo, err := os.Create(newpath)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s", newpath)
	}
	defer writeTo.Close()

	fmt.Fprintf(writeTo, "---\n%s---\n%s", string(yml), string(output))
	return nil
}

// yaml front matter と本文を切り離す
func splitMarkdown(content []rune) (yml []byte, body []rune) {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	if !scanner.Scan() {
		return nil, nil
	}

	// --- が見つからなかったら front matter なし
	if scanner.Text() != "---" {
		return nil, content
	}

	// --- が見つかるまで front matter に追加していく
	frontMatter := make([]rune, 0)
	endFound := false
	for scanner.Scan() {
		if scanner.Text() == "---" {
			endFound = true
			break
		} else {
			frontMatter = append(frontMatter, []rune(scanner.Text())...)
			frontMatter = append(frontMatter, '\n')
		}
	}

	if !endFound {
		return nil, content
	}

	for scanner.Scan() {
		body = append(body, []rune(scanner.Text())...)
		body = append(body, '\n')
	}
	return []byte(string(frontMatter)), body
}
