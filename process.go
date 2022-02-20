package main

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
	"github.com/qawatake/obsdconv/process"
)

type processorImplWithErrHandling struct {
	debug  bool
	sub    process.Processor
	errbuf []error
}

func newProcessorImplWithErrHandling(debug bool, subprocessor process.Processor) *processorImplWithErrHandling {
	return &processorImplWithErrHandling{
		debug: debug,
		sub:   subprocessor,
	}
}

func (p *processorImplWithErrHandling) Process(relativePath, orgpath, newpath string) error {
	err := p.sub.Process(relativePath, orgpath, newpath)

	if err == nil {
		return nil
	}

	// 予想済みのエラーの場合は処理を止めずに, エラー出力だけする
	public, debug, buffered := handleErr(orgpath, err)
	if public == nil && debug == nil {
		if buffered != nil {
			p.errbuf = append(p.errbuf, buffered)
		}
		return nil
	}

	if p.debug {
		return debug
	} else {
		return public
	}
}

func newDefaultProcessor(config *configuration) (processor *processorImplWithErrHandling, err error) {
	skipper, err := process.NewSkipper(filepath.Join(config.src, DEFAULT_IGNORE_FILE_NAME))
	if err != nil {
		return nil, err
	}
	db := process.WrapForSkipping(convert.NewPathDB(config.src), skipper)

	if config.strictref {
		db = convert.WrapForReturningNotFoundPathError(db)
	}

	bc := newBodyConverterImpl(db, config.cptag || config.synctag, config.rmtag, config.cmmt, config.title || config.alias || config.synctlal, config.link, config.rmH1, config.baseUrl)
	remap, err := parseRemap(config.remapkey)
	if err != nil {
		return nil, err
	}
	yc := newYamlConverterImpl(config.synctag, config.synctlal, config.publishable, remap)
	passer := newArgPasserImpl(config.title || config.synctlal, config.alias || config.synctlal)
	examinator := newYamlExaminatorImpl(config.filter, config.publishable)
	return newProcessorImplWithErrHandling(config.debug, process.NewProcessor(bc, yc, passer, examinator)), nil
}

func handleErr(path string, err error) (public error, debug error, buffered error) {
	orgErr := errors.Cause(err)
	e, ok := orgErr.(convert.ErrConvert)
	if !ok {
		e := fmt.Errorf("[FATAL] path: %s | %v", path, err)
		return e, e, nil
	}

	line := e.Line()
	ee, ok := errors.Cause(e.Source()).(convert.ErrTransform)
	if !ok {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | cause of source of ErrConvert does not implement ErrTransform: ErrConvert: %w", path, line, e)
		return public, debug, nil
	}

	if ee.Kind() == convert.ERR_KIND_UNEXPECTED {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | undefined kind of ErrTransform: ErrTransform: %w", path, line, ee)
		return public, debug, nil
	}

	// 想定済みのエラー
	return nil, nil, errors.Wrapf(ee, "[ERROR] path: %s, around line: %d", path, line)
}
