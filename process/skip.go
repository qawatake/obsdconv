package process

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/text/unicode/norm"
)

type Skipper interface {
	Skip(path string) bool
}

type skipperImpl struct {
	skipdict map[string]bool
}

func (s *skipperImpl) Skip(path string) bool {
	return s.skipdict[filepath.ToSlash(norm.NFC.String(path))]
}

func (s *skipperImpl) add(path string) {
	s.skipdict[filepath.ToSlash(norm.NFC.String(path))] = true
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
	if scanner.Scan() {
		skipper.add(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "error occurred during scanning %s", skipsource)
	}
	return skipper, nil
}
