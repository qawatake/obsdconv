---
aliases:
- existing-alias
tags:
- existing-tag
---
# sample file for -std -rmh1 << #will_be_removed_in_title_and_alias >>
↑ this H1 will be removed.

## Copy tags
Tags will be copied to `tags` field in front matter.
<< #obsidian >> <- this tag will be copied (and be removed).

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
%%
	#comment-block
%%
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
<< #will_remain >> <- this tag will be removed

## Convert Links
### Internal Links
[[blank]]

### Obsidian URL
[obsidian url](obsidian://open?vault=obsidian&file=blank)

### Embeds
![[image.png]]

## Remove Obsidian Comment Blocks
%%
This comment block will be removed.
%%
