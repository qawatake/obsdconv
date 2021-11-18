package convert

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/qawatake/obsdconv/scan"
)

type TransformerFunc func(raw []rune, ptr int) (advance int, tobewritten []rune, err error)

type Converter struct {
	transformers []TransformerFunc
}

func (c *Converter) Set(t TransformerFunc) {
	c.transformers = append(c.transformers, t)
}

func (c *Converter) Convert(raw []rune) (output []rune, err error) {
	output = make([]rune, 0)
	ptr := 0
	for ptr < len(raw) {
		org := ptr
		for _, transformer := range c.transformers {
			advance, tobewritten, err := transformer(raw, ptr)
			if err != nil {
				err := newErrConvert(err)
				err.SetLine(currentLine(raw, ptr))
				return nil, errors.Wrap(err, "transformation failed")
			}
			if advance > 0 {
				output = append(output, tobewritten...)
				ptr += advance
				break
			}
		}
		if ptr <= org {
			err := newErrConvert(fmt.Errorf("caught by no transformer"))
			err.SetLine(currentLine(raw, ptr))
			return nil, err
		}
	}

	return output, nil
}

type ScannerFunc func(raw []rune, ptr int) (advance int)

func MiddlewareAsIs(scanner ScannerFunc) TransformerFunc {
	return TransformerFunc(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scanner(raw, ptr)
		return advance, raw[ptr : ptr+advance], nil
	})
}

func NewTagRemover() *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance, nil, nil
	})
	c.Set(TransformNone)
	return c
}

func NewTagFinder(tags map[string]struct{}) *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, t := scan.ScanTag(raw, ptr)
		if advance > 0 {
			tags[t] = struct{}{}
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(TransformNone)
	return c
}

func NewTitleFinder(title *string) *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, level, headertext := scan.ScanHeader(raw, ptr)
		if level == 1 && *title == "" {
			*title = headertext
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func newLinkConverter(internal, embeds, external TransformerFunc) *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(external)
	c.Set(internal)
	c.Set(embeds)
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func NewLinkConverter(db PathDB) *Converter {
	internal := defaultTransformInternalLinkFunc(db)
	embeds := defaultTransformEmbedsFunc(db)
	external := defaultTransformExternalLinkFunc(db)
	return newLinkConverter(internal, embeds, external)
}

func NewCommentEraser() *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanComment(raw, ptr)
		if advance == 0 {
			return 0, nil, nil
		}
		return advance, nil, nil
	})
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func NewInternalLinkPlainConverter() *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(TransformInternalLinkToPlain)
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func NewH1Remover() *Converter {
	c := new(Converter)

	c.Set(MiddlewareAsIs(scan.ScanEscaped))
	c.Set(MiddlewareAsIs(scan.ScanCodeBlock))
	c.Set(MiddlewareAsIs(scan.ScanComment))
	c.Set(MiddlewareAsIs(scan.ScanMathBlock))
	c.Set(MiddlewareAsIs(scan.ScanNormalComment))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(MiddlewareAsIs(scan.ScanInlineMath))
	c.Set(MiddlewareAsIs(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, level, _ := scan.ScanHeader(raw, ptr)
		if level == 1 {
			return advance, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(MiddlewareAsIs(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}
