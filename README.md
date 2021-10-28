# obsdconv
CLI program and a Go package to convert [Obsidian](https://obsidian.md/) files in several ways.
You can use the program both for exporting Obsidian files to static site generators (e.g., Hugo), and for modifying front matters.

Obsdconv enables you to
- remove tags from text,
- copy tags in text to front matter `tags` field,
- set H1 content to front matter `title` and `aliases` field,
- convert internal links `[[file]]` , embeds `![[image]]`, and external links with Obsidian URI `[text](obsidian://open?vault=notes&file=filename)` to the standard format, and etc.

## Installation
We provide binaries for several platforms.
Please download the one suitable to your environment.
Or if you have a go runtime, you can build a binary by running
`go mod tidy && go build`, to get `obsdconv`.
Remenber to set `PATH` for the binary.

## Quick Start
Run
```bash
obsdconv -src src -dst dst -std
```
where `src` is a directory with Obsidian files or an Obsidian file, `dst` is a directory to which processed files will be exported.
With `std` flag, obsdconv exports Obsidian files in the standard format.
That is, obsdconv
- removes tags from text,
- copy tags in text to front matter,
- copy H1 content to `title` and `aliases` fields,
- remove comment blocks,
- convert internal links, embeds, and Obsidian URI's,
- set front matter `draft` to be consistent with `publish`

See `sample` directory.
We can get `sample/ouput.md` from `sample/input.md` by running `obsdconv -src sample -dst sample -std` (and renaming the generated file).

## Options
Available options are as follows:

flag | meaning | \*
--- | --- | ---
`src` | directory containing Obsidian files  | **required**
`dst` | destination to which generated files located | **required**
`rmtag` | remove tags from text | optional
`cptag` | copy tags from text to `tags` field in front matter | optional
`title` | copy H1 content to `title` field in front matter | optional
`alias` | copy H1 content to `aliases` field in front matter | optional
`link` | convert internal links, embeds, and Obsidian URI in the standart format | optional
`cmmt` | remove comment blocks | optional
`pub` | `publish: true` -> `draft: false`, `publish: false` -> `draft: true`, no `publish` field -> draft: true. If `draft` field explicityly specified, then leave it as is. | optional
`obs` | = `-cptag -title -alias` | optional
`std` | = `-cptag -title -alias -rmtag -link -cmmt -pub` | optional

Note that
- individual flag overrides `obs` and `std`.
That is, if you specify `-title=0` and `-obs`, `-title=0` wins and `title` field will not copied from H1 content.
- if `src` = `dst`, then original files will be overwritten. Be careful!!


## Customization
The main program is based on two packages `github.com/qawatake/obsdconv/scan` and `github.com/qawatake/obsdconv/convert`.
If you would like to customize the program,
- combine existing `Converters`, or
- create new scanning functions to get a new `Converter`.

Please request new features that will be commonly used.
