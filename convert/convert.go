package convert

import (
	"fmt"

	"github.com/qawatake/obsd2hugo/scan"
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
				return nil, fmt.Errorf("transformation failed: %w", err)
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

type Middleware func(ScannerFunc) TransformerFunc

func DefaultMiddleware(scanner ScannerFunc) TransformerFunc {
	return TransformerFunc(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scanner(raw, ptr)
		return advance, raw[ptr : ptr+advance], nil
	})
}

func NewTagRemover() *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scan.ScanEscaped))
	c.Set(DefaultMiddleware(scan.ScanCodeBlock))
	c.Set(DefaultMiddleware(scan.ScanComment))
	c.Set(DefaultMiddleware(scan.ScanMathBlock))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(scan.ScanInlineMath))
	c.Set(DefaultMiddleware(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(TransformTag)
	c.Set(TransformNone)
	return c
}

func NewTagFinder(tags map[string]struct{}) *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scan.ScanEscaped))
	c.Set(DefaultMiddleware(scan.ScanCodeBlock))
	c.Set(DefaultMiddleware(scan.ScanComment))
	c.Set(DefaultMiddleware(scan.ScanMathBlock))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(scan.ScanInlineMath))
	c.Set(DefaultMiddleware(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
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

	c.Set(DefaultMiddleware(scan.ScanEscaped))
	c.Set(DefaultMiddleware(scan.ScanCodeBlock))
	c.Set(DefaultMiddleware(scan.ScanComment))
	c.Set(DefaultMiddleware(scan.ScanMathBlock))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(scan.ScanInlineMath))
	c.Set(DefaultMiddleware(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance, level, headertext := scan.ScanHeader(raw, ptr)
		if level == 1 && *title == "" {
			*title = headertext
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func NewLinkConverter(vault string) *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scan.ScanEscaped))
	c.Set(DefaultMiddleware(scan.ScanCodeBlock))
	c.Set(DefaultMiddleware(scan.ScanComment))
	c.Set(DefaultMiddleware(scan.ScanMathBlock))
	c.Set(TransformExternalLinkFunc(vault))
	c.Set(TransformInternalLinkFunc(vault))
	c.Set(TransformEmbedsFunc(vault))
	c.Set(DefaultMiddleware(scan.ScanInlineMath))
	c.Set(DefaultMiddleware(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}

func NewCommentEraser() *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scan.ScanEscaped))
	c.Set(DefaultMiddleware(scan.ScanCodeBlock))
	c.Set(TransformComment)
	c.Set(DefaultMiddleware(scan.ScanMathBlock))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _, _ = scan.ScanExternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanInternalLink(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanEmbeds(raw, ptr)
		return advance
	}))
	c.Set(DefaultMiddleware(scan.ScanInlineMath))
	c.Set(DefaultMiddleware(scan.ScanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune, err error) {
		advance = scan.ScanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil, nil
		}
		return advance, raw[ptr : ptr+advance], nil
	})
	c.Set(DefaultMiddleware(func(raw []rune, ptr int) (advance int) {
		advance, _ = scan.ScanTag(raw, ptr)
		return advance
	}))
	c.Set(TransformNone)
	return c
}
