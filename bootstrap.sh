#!/bin/bash

set -e

if !command -v yq &> /dev/null; then
  echo "yq needs to be installed"
  exit 2
fi
if !command -v docker &> /dev/null; then
  echo "docker needs to be installed"
  exit 2
fi
if [ -z "$1" ]; then
  echo "You must specify a name for the bootstrapped image, this should be in the format registry"
  exit 2
fi

echo "Making temp dir"
DIR=$(mktemp -d)

#Get latest version info
MIRROR="https://uk.alpinelinux.org/alpine/"
RELBASE="$MIRROR/latest-stable/releases/x86_64"
YAML=$(curl -qSs $RELBASE/latest-releases.yaml | yq -r '.[] | select(.title=="Mini root filesystem")')
FILE=$(echo $YAML | yq -r '.file')
SHA=$(echo $YAML | yq -r '.sha256')

#Download and verify file
curl -qSs -o $DIR/$FILE $RELBASE/$FILE

if [[ "$(echo "$SHA *$DIR/$FILE" | sha256sum -c --status -)" -ne "0" ]]; then
  echo "Downloaded filesystem checksum does not match, exiting."
  exit
fi

mkdir -p $DIR/fs
tar -C $DIR/fs -xzf $DIR/$FILE;

echo "[advice]" > $DIR/fs/etc/gitconfig
echo "    detachedHead = false" >> $DIR/fs/etc/gitconfig

#Create Dockerfile
echo "FROM scratch" > $DIR/Dockerfile
echo "ADD /fs /" >> $DIR/Dockerfile
echo "CMD [\"/bin/sh\"]" >> $DIR/Dockerfile

cat $DIR/Dockerfile

#Build and push image
docker build -t $1/alpine $DIR
docker push $1/alpine

#Removing temp dir
rm -rf $DIR

go run github.com/csmith/contempt/cmd/contempt@latest -force-build=1 -push=1 -commit=1 -registry $1 -source-link "https://github.com/greboid/dockerfiles/blob/master/" . .
