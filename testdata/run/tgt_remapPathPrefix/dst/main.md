---
aliases:
- existing-alias
tags:
- existing-tag
---
# H1 << #in_title >> [external link](https://example.com)

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
[blank](posts/blank)

#### with display name
[with display name](posts/blank)

#### with fragments
[blank > section > subsection](posts/blank#subsection)

#### only fragments
[section](posts/main#section)

### External Link
#### tag in display name
[tag in display name #tag_in_display_name_of_external_link](https://example.com)

#### file id
[file id](posts/blank)

#### file id with fragments
[file id with fragments](posts/blank#section)

#### obsidian url
[obsidian url](posts/blank)

#### shorthand obsidian url
[shorthand obsidian url](posts/blank)

#### variant of external link
[variant of external link #tag_in_display_name_of_var_external_link][variant #variant]

[variant #variant]:https://example.com

### Embeds
![image.svg](posts/image.svg)

## Obsidian Comment Block
%%
This comment block will be removed.
%%