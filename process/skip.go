package process

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
	"golang.org/x/text/unicode/norm"
)

type Skipper interface {
	Skip(path string) (tobeskipped bool)
}

type skipperImpl struct {
	skipdict map[string]bool
}

func (s *skipperImpl) Skip(path string) bool {
	// check parent directories one by one
	// if abc/def is registered, then abc/def/ghi.md will be skipped.
	cur := filepath.ToSlash(norm.NFC.String(path))
	for cur != "." {
		if s.skipdict[cur] {
			return true
		}
		cur = filepath.Dir(cur)
	}
	return false
}

func (s *skipperImpl) add(path string) {
	path = norm.NFC.String(path)
	path = filepath.ToSlash(path)
	path = filepath.Clean(path)
	s.skipdict[path] = true
}

func NewSkipper(skipsource string) (Skipper, error) {
	skipper := new(skipperImpl)
	skipper.skipdict = make(map[string]bool)

	file, err := os.Open(skipsource)
	if err != nil {
		if os.IsNotExist(err) {
			return skipper, nil
		}
		return nil, errors.Wrapf(err, "failed to open %s", skipsource)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		skipper.add(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "error occurred during scanning %s", skipsource)
	}
	return skipper, nil
}

type pathDBWrapperImplSkipping struct {
	original convert.PathDB
	skipper  Skipper
}

func (w *pathDBWrapperImplSkipping) Get(fileId string) (path string, err error) {
	if w.original == nil {
		panic("original PathDB not set but used")
	}
	if w.skipper == nil {
		panic("skipper not set but used")
	}
	path, err = w.original.Get(fileId)
	if err != nil {
		return "", err
	}
	if w.skipper.Skip(path) {
		return "", err
	}
	return path, nil
}

func WrapForSkipping(original convert.PathDB, skipper Skipper) convert.PathDB {
	return &pathDBWrapperImplSkipping{
		original: original,
		skipper:  skipper,
	}
}
