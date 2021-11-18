package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/qawatake/obsdconv/convert"
	"github.com/qawatake/obsdconv/process"
)

type processorImplWithErrHandling struct {
	debug bool
	sub   process.Processor
}

func newProcessorImplWithErrHandling(debug bool, subprocessor process.Processor) *processorImplWithErrHandling {
	return &processorImplWithErrHandling{
		debug: debug,
		sub:   subprocessor,
	}
}

func (p *processorImplWithErrHandling) Process(orgpath, newpath string) error {
	err := p.sub.Process(orgpath, newpath)

	if err == nil {
		return nil
	}

	// 予想済みのエラーの場合は処理を止めずに, エラー出力だけする
	public, debug := handleErr(orgpath, err)
	if public == nil && debug == nil {
		return nil
	}

	if p.debug {
		return debug
	} else {
		return public
	}
}

func newDefaultProcessor(flags *flagBundle) process.Processor {
	db := convert.NewPathDB(flags.src)
	bc := newBodyConverterImpl(db, flags.cptag, flags.rmtag, flags.cmmt, flags.title, flags.link, flags.rmH1)
	yc := newYamlConverterImpl(flags.publishable)
	passer := argPasserFunc(passArg)
	examinator := newYamlExaminatorImpl(flags.publishable)
	return newProcessorImplWithErrHandling(flags.debug, process.NewProcessor(bc, yc, passer, examinator))
}

func handleErr(path string, err error) (public error, debug error) {
	orgErr := errors.Cause(err)
	e, ok := orgErr.(convert.ErrConvert)
	if !ok {
		e := fmt.Errorf("[FATAL] path: %s | %v", path, err)
		return e, e
	}

	line := e.Line()
	ee, ok := errors.Cause(e.Source()).(convert.ErrTransform)
	if !ok {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | cause of source of ErrConvert does not implement ErrTransform: ErrConvert: %w", path, line, e)
		return public, debug
	}

	if ee.Kind() == convert.ERR_KIND_UNEXPECTED {
		public = fmt.Errorf("[FATAL] path: %s, around line: %d | failed to convert", path, line)
		debug = fmt.Errorf("[FATAL] path: %s, around line: %d | undefined kind of ErrTransform: ErrTransform: %w", path, line, ee)
		return public, debug
	}

	// 想定済みのエラー
	fmt.Fprintf(os.Stderr, "[ERROR] path: %s, around line: %d | %v\n", path, line, ee)
	return nil, nil
}
