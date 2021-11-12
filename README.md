# obsdconv
[![test](https://github.com/qawatake/obsdconv/actions/workflows/test.yml/badge.svg)](https://github.com/qawatake/obsdconv/actions/workflows/test.yml)

CLI program and a Go package to convert [Obsidian](https://obsidian.md/) files in multiple ways.
You can use the program both for exporting Obsidian files to static site generators (e.g., Hugo), and for modifying front matters.

Obsdconv enables you to
- remove tags from text,
- copy tags in text to front matter `tags` field,
- set H1 content to front matter `title` and `aliases` field,
- convert internal links `[[file]]` , embeds `![[image]]`, and external links with Obsidian URI `[text](obsidian://open?vault=notes&file=filename)` to the standard format, and etc.

## Installation
We provide binaries for multiple platforms.
Please download the one suitable to your environment.
Or if you have a go runtime, you can build a binary by running
`go mod tidy && go build`, to get `obsdconv`.

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
`pub` | convert only files with `publish: true` or `draft: false`. For files with `publish: true`, add `draft: false`. | optional
`rmh1` | remove H1 | optional
`obs` | = `-cptag -title -alias` | optional
`std` | = `-cptag -title -alias -rmtag -link -cmmt -pub` | optional
`verion` | display the version currently installed | optional
`debug` | display error messages for developers | optional

Note that
- individual flag overrides `obs` and `std`.
That is, if you specify `-title=0` and `-obs`, `-title=0` wins and `title` field will not copied from H1 content.
- if `src` = `dst`, then original files will be overwritten. Be careful!!

## Ignore Files
You can ignore paths by specifying them in a file named `.obsdconvignore`.
Put `.obsdconvignore` in `src` directory and write a path in each line like this:
```.obsdconvignore
.obsdconvignore
static/private/
notes/private/
notes/mycredential.md
```
- By default, non-markdown files will be copied to `dst` directory.
