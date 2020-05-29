#!/usr/bin/env bash

modifyFiles=$(git status | grep -e '\.proto$' | grep -e 'modified:' | sed -e "s/modified://" )
newProtoFiles=$(git status | grep -e 'new file:' | grep -e '\.proto$' | sed -e "s/new file://" )
if [[ -z "$modifyProtoFiles" && -z "$newProtoFiles" ]]
then
    echo
else
    echo -----clang-format-----
        #-------------------------
        for f1 in ${modifyFiles[@]}; do
            echo $f1
            l1=$(git blame -p $f1 | grep "0000000000000000000000000000000000000000" | awk -F' ' '{if ($4 > 0) {printf "-lines=%d:%d\n", $2, $2+$4}}')
            for line in "${l1[@]}"; do
                clang-format -i $f1 $l1
            done
        done
        #-------------------------
        for f2 in "${newProtoFiles[@]}"; do
            echo $f2
            clang-format -i $f2
        done
        #-------------------------
    echo
fi

files=$(git status | grep -e 'modified:' -e 'new file:' | grep '\.go' | awk -F ':' '{print $2}')
if [ -z "$files" ]
then
    echo
else
    echo ----gofmt----
    gofmt -w -l $files
    echo
fi

