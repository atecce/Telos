#!/bin/sh

GOOS=linux GOARCH=386 go build -o $1/$1 github.com/atecce/investigations/$1
docker build -t atec/$1 $1