#!/bin/bash
mkdir tmp
cp -r sample/. tmp
go run . -src tmp -dst tmp -std
diff sample/output.md tmp/obsidian.md
rm -r tmp