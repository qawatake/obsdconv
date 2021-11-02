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
%%
	#comment-block
%%
- ↑ this comment block will disappear.

### Normal Comment Block
<!--
	#normal-comment-block
-->

### Inlines
- `#inline-code`
- $#inline-math$

## H1 -> Title
H1 content will be copied to `title` field in front matter.

## H1 -> Aliases
H1 content will be copied to `aliases` field in front matter.

## Internal Links
- simple: [[ test ]]
- with display name: [[ test | DISPLAY NAME ]]
- with fragments: [[ test#section ]]

## Embeds
- simple: ![[ image.png ]]
- with alt: ![[ image.png | ALT TEXT ]]

## External Links
- Obsidian URI: [ obsidian uri ](obsidian://open?vault=obsidian&file=test)
- extension omitted: [ noext ](test)