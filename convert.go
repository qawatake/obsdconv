package main

import (
	"log"
)

type TransformerFunc func(raw []rune, ptr int) (advance int, tobewritten []rune)

type Converter struct {
	transformers []TransformerFunc
}

func (c *Converter) Set(t TransformerFunc) {
	c.transformers = append(c.transformers, t)
}

func (c *Converter) Convert(raw []rune) (output []rune) {
	output = make([]rune, 0)
	ptr := 0
	for ptr < len(raw) {
		org := ptr
		for _, scanner := range c.transformers {
			advance, tobewritten := scanner(raw, ptr)
			if advance > 0 {
				output = append(output, tobewritten...)
				ptr += advance
				break
			}
		}
		if ptr <= org {
			log.Fatal("pointer did not proceed")
		}
	}
	return output
}

type ScannerFunc func(raw []rune, ptr int) (advance int)

type Middleware func(ScannerFunc) TransformerFunc

func DefaultMiddleware(scanner ScannerFunc) TransformerFunc {
	return TransformerFunc(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanner(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
}

func NewDefaultConverter(vault string) *Converter {
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
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformTag)
	c.Set(TransformNone)
	return c
}

func NewTagRemover() *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	// c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanComment))
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _, _ = scanExternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	// c.Set(TransformExternalLinkFunc(vault))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanInternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	// c.Set(TransformInternalLinkFunc(vault))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanEmbeds(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	// c.Set(TransformEmbedsFunc(vault))
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformTag)
	c.Set(TransformNone)
	return c
}

func NewTagFinder(tags map[string]struct{}) *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	// c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanComment))
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _, _ = scanExternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanInternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanEmbeds(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, t := scanTag(raw, ptr)
		if advance > 0 {
			tags[t] = struct{}{}
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformNone)
	return c
}

func NewTitleFinder(title *string) *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	// c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanComment))
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _, _ = scanExternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanInternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanEmbeds(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, level, headertext := scanHeader(raw, ptr)
		if level == 1 && *title == "" {
			*title = headertext
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanTag(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformNone)
	return c
}

func NewLinkConverter(vault string) *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	// c.Set(TransformComment)
	c.Set(DefaultMiddleware(scanComment))
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(TransformExternalLinkFunc(vault))
	// c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
	// 	advance, _, _ = scanExternalLink(raw, ptr)
	// 	return advance, raw[ptr : ptr+advance]
	// })
	c.Set(TransformInternalLinkFunc(vault))
	// c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
	// 	advance, _ = scanInternalLink(raw, ptr)
	// 	return advance, raw[ptr : ptr+advance]
	// })
	c.Set(TransformEmbedsFunc(vault))
	// c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
	// 	advance, _ = scanEmbeds(raw, ptr)
	// 	return advance, raw[ptr : ptr+advance]
	// })
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	// c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
	// 	advance, level, headertext := scanHeader(raw, ptr)
	// 	if level == 1 && *title == "" {
	// 		*title = headertext
	// 	}
	// 	return advance, raw[ptr : ptr+advance]
	// })
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanTag(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformNone)
	return c
}

func NewCommentEraser() *Converter {
	c := new(Converter)

	c.Set(DefaultMiddleware(scanEscaped))
	c.Set(DefaultMiddleware(scanCodeBlock))
	c.Set(TransformComment)
	// c.Set(DefaultMiddleware(scanComment))
	c.Set(DefaultMiddleware(scanMathBlock))
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _, _ = scanExternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanInternalLink(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanEmbeds(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(DefaultMiddleware(scanInlineMath))
	c.Set(DefaultMiddleware(scanInlineCode))
	// c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
	// 	advance, level, headertext := scanHeader(raw, ptr)
	// 	if level == 1 && *title == "" {
	// 		*title = headertext
	// 	}
	// 	return advance, raw[ptr : ptr+advance]
	// })
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance = scanRepeat(raw, ptr, "#")
		if advance <= 1 {
			return 0, nil
		}
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(func(raw []rune, ptr int) (advance int, tobewritten []rune) {
		advance, _ = scanTag(raw, ptr)
		return advance, raw[ptr : ptr+advance]
	})
	c.Set(TransformNone)
	return c
}
