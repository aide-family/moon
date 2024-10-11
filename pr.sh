#!/bin/bash
git fetch origin pull/$1/head:pr/$1 && git checkout pr/$1 --no-pager 2> /dev/null || git branch -r | grep $1
git checkout pr/$1