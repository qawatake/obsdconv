---
aliases:
- existing-alias
- sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref) <<  >>
tags:
- existing-tag
- obsidian
- will_be_removed_in_title_and_alias
- will_remain
title: sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref)
  <<  >>
---
# sample file for -std (= -cptag -rmtag -title -alias -link -cmmt -strictref) <<  >>

## Copy tags
Tags will be copied to `tags` field in front matter.
<<  >> <- this tag will be copied (and be removed).

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
<<  >> <- this tag will be removed

## Convert Links
### Internal Links
[blank](blank.md)

### Obsidian URL
[obsidian url](blank.md)

### Embeds
![image.png](image.png)

## Remove Obsidian Comment Blocks
