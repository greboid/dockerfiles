#!/bin/sh

if !command -v yq &> /dev/null; then
  echo "docker needs to be installed"
  exit 2
fi
if !command -v docker &> /dev/null; then
  echo "docker needs to be installed"
  exit 2
fi
if [ -z "$1" ]; then
  echo "You must specify a name for the bootstrapped image, this should be in the format registry/image"
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
if [ "$(echo "$SHA *$FILE" | sha256sum -c --status -)" -ne "0" ]; then
  echo "Filesystem incorrect"
fi
echo "File verifies, next"

#Create Dockerfile
echo "FROM scratch" > $DIR/Dockerfile
echo "ADD $FILE /" >> $DIR/Dockerfile
echo "CMD [\"/bin/sh\"]" >> $DIR/Dockerfile

#Build and push image
docker build -t $1 $DIR
docker push $1

#Removing temp dir
rm -rf $DIR
