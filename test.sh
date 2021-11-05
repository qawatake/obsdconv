#!/bin/bash
rm -rf tmp
mkdir tmp
cp -r sample/. tmp
go run . -src tmp -dst tmp -std
diff sample/output.md tmp/input.md
result=$?
if [ $result -eq 0 ]; then
  echo OK
  rm -rf tmp
else
  echo FAIL
fi