#!/bin/bash
mkdir tmp
cp -r sample/. tmp
go run . -src tmp -dst tmp -std
diff sample/output.md tmp/input.md
rm -r tmp