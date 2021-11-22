---
aliases:
- existing-alias
- sample file for -std -rmh1 <<  >>
tags:
- existing-tag
- obsidian
- will_be_removed_from_text
- will_be_removed_in_title_and_alias
title: sample file for -std -rmh1 <<  >>
---
↑ this H1 will be removed.

## Copy tags
Tags will be copied to `tags` field in front matter.
<<  >> <- `#obsidian` will be copied (and be removed).

### Not tags
Tags are escaped in the following.

#### Code Block
```
	#code-block
```

#### Math Block
$$
	#math-block
$$

#### Comment Block

(↑ this comment block will be removed)

#### Inline Code
`#inline-code`

#### Inline Math
$#inline-math$


## Set titles
H1 content will be copied to `title` field in front matter.
In this case,
- tags are removed,
- internal links and external links are converted to display names only.

## Set aliases
H1 content will be copied to `aliases` field in front matter.
H1 content will be processed like `title`.

### Remove Tags
<<  >> <- `#will_be_removed_from_text` will be removed

## Convert Links
### Internal Links
[blank](blank.md)

### Obsidian URL
[obsidian url](blank.md)

### Embeds
![image.svg](image.svg)

## Remove Obsidian Comment Blocks
