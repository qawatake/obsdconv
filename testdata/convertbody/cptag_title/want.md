# test source file #test
Let's compare `obsidian.md` and `output.md`

## Tags
Tags will be copied to `tags` field in front matter and will be removed. #obsidian <- there is a tag in src.

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
%%
	#comment-block
%%
- ↑ this comment block will disappear.

### Inlines
- `#inline-code`
- $#inline-math$

## H1 -> Title
H1 content will be copied to `title` field in front matter.

## H1 -> Aliases
H1 content will be copied to `aliases` field in front matter.

## Internal Links
- simple: [[ output ]]
- with display name: [[ output | DISPLAY NAME ]]
- with fragments: [[ output#section ]]

## Embeds
- simple: ![[ output.md ]]
- with alt: ![[ output.md | ALT TEXT ]]
	- in the current version, we do not expanding embedded markdown notes.

## External Links
- Obsidian URI: [ obsidian uri ](obsidian://open?vault=obsidian&file=output)
	- we support `open` action only.
- extension omitted: [ noext ](output)

## `publish` field in front matter
Since `publish: true`, we will have `draft: false` in front matter.