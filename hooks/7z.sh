#!/bin/sh

mkdir tmp
7z x "$1" -otmp
rsync -avz tmp/* jade:/mnt/music
rm -rf tmp
