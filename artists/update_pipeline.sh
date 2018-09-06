#!/bin/sh

GOOS=linux GOARCH=386 go build
docker build -t atec/artists .
pachctl update-pipeline --reprocess --push-images -f artists.json --password uxR-ymA-uT5-YNh -u atec