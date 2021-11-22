---
aliases:
- existing-alias
- sample file for -obs (= -cptag -title -alias) <<  >>
tags:
- existing-tag
- obsidian
- will_be_removed_in_title_and_alias
- will_remain
title: sample file for -obs (= -cptag -title -alias) <<  >>
---
# sample file for -obs (= -cptag -title -alias) << #will_be_removed_in_title_and_alias >>

## Copy tags
Tags will be copied to `tags` field in front matter.
<< #obsidian >> <- this tag will be copied.

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

## Obsidian formats uneffected
The following formats will be unchanged.

### Tags
#will_remain

### Internal Links
[[blank]]

### Obsidian URL
[obsidian url](obsidian://open?vault=obsidian&file=blank)

### Embeds
![[image.png]]
