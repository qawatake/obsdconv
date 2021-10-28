---
aliases:
- existing-alias
- when std flag applied
draft: false
publish: true
tags:
- existing-tag
- obsidian
title: when std flag applied
---
# when std flag applied
Let's compare `obsidian.md` and `output.md`

## Tags
Tags will be copied to `tags` field in front matter and will be removed.
- 
- ↑ this tag will disappear.

## Not tags
Tags are escaped in the following.

### Code Block
```
	#code-block
```

### Math Block
$$
	#math-block
$$

### Comment Block

- ↑ this comment block will disappear.

### Inlines
- `#inline-code`
- $#inline-math$

## H1 -> Title
H1 content will be copied to `title` field in front matter.

## H1 -> Aliases
H1 content will be copied to `aliases` field in front matter.

## Internal Links
- simple: [output](output.md)
- with display name: [DISPLAY NAME](output.md)
- with fragments: [output > section](output.md#section)

## Embeds
- simple: ![output.md](output.md)
- with alt: ![ALT TEXT](output.md)
	- in the current version, we do not expanding embedded markdown notes.

## External Links
- Obsidian URI: [obsidian uri](output.md)
	- we support `open` action only.
- extension omitted: [noext](output.md)

## `publish` field in front matter
Since `publish: true`, we will have `draft: false` in front matter.
