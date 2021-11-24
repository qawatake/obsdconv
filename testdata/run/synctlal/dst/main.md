---
aliases:
- existing-alias
- H1 <<  >> external link internal link
tags:
- existing-tag
title: H1 <<  >> external link internal link
---
# H1 << #in_title >> [external link](https://example.com) [[internal_link | internal link]]

## Tags
<< #obsidian >>

### Not tags
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

## Links

### Internal Link
#### simple
[[blank]]

#### with display name
[[blank | with display name]]

#### with fragments
[[blank#section#subsection]]

### External Link
#### tag in display name
[tag in display name #tag_in_display_name_of_external_link](https://example.com)

#### file id
[file id](blank)

#### file id with fragments
[file id with fragments](blank#section)

#### obsidian url
[obsidian url](obsidian://open?vault=obsidian&file=blank)

#### shorthand obsidian url
[shorthand obsidian url](obsidian://vault/my_vault/blank)

#### variant of external link
[variant of external link #tag_in_display_name_of_var_external_link][variant #variant]

[variant #variant]:https://example.com

### Embeds
![[image.svg]]

## Obsidian Comment Block
%%
This comment block will be removed.
%%
