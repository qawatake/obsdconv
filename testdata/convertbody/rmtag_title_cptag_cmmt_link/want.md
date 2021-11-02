# test source file <<>>
Let's compare `obsidian.md` and `output.md`

## Tags
Tags will be copied to `tags` field in front matter and will be removed. <<>> <- there is a tag in src.

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
- simple: [test](test.md)
- with display name: [DISPLAY NAME](test.md)
- with fragments: [test > section](test.md#section)

## Embeds
- simple: ![image.png](image.png)
- with alt: ![ALT TEXT](image.png)

## External Links
- Obsidian URI: [obsidian uri](test.md)
- extension omitted: [noext](test.md)