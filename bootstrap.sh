#!/bin/sh

if !command -v jq &> /dev/null; then
  echo "jq needs to be installed"
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

#This is the base API path of the alpine/apk-tools repo on gitlab
GL="https://gitlab.alpinelinux.org/api/v4/projects/5"
#This gets the latest version of apk-tools
VER=$(curl -s $GL/releases/ | jq '.[]' | jq -r '.name' | head -1)
#This is the URL to download the static apk build
DL="$GL/packages/generic/$VER/x86_64/apk.static"

echo "Making temp dir"
DIR=$(mktemp -d)

echo "Downloading static apk"
curl -qSs -o $DIR/apk.static $DL
chmod +x $DIR/apk.static

#This script does
#  - Adds edge repos and installs buildah
#  - Bootstraps a new alpine into /alpine
#  - Builds a base alpine image
#  - Pushes the base alpine image
cat << EOF > $DIR/run.sh
#!/bin/sh
echo "https://mirrors.melbourne.co.uk/alpine/edge/main" > /etc/apk/repositories
echo "https://mirrors.melbourne.co.uk/alpine/edge/community" >> /etc/apk/repositories
apk add buildah

/apk.static -X https://mirrors.melbourne.co.uk/alpine/latest-stable/main --allow-untrusted -p /alpine --initdb add alpine-base

buildah bud -t $1 -f /Dockerfile
buildah push --authfile /config.json $1
EOF
chmod +x $DIR/run.sh

#This dockerfile is used in the container to create the base image
cat << EOF > $DIR/Dockerfile2
FROM scratch
COPY /alpine /
EOF

#This dockerfile bootstraps an alpine container so it can build the base image
cat << EOF > $DIR/Dockerfile
FROM scratch
COPY apk.static /apk.static
COPY run.sh /run.sh
COPY Dockerfile2 /Dockerfile
RUN ["/apk.static", "-X", "http://mirrors.melbourne.co.uk/alpine/latest-stable/main", "--allow-untrusted", "-p", "/", "--initdb", "add", "alpine-base"]
CMD ["/run.sh"]
EOF

echo "Building bootstrap"
docker build -t bootstrap -f $DIR/Dockerfile $DIR

echo "Running bootstrap"
docker run --rm --privileged -v ~/.docker/config.json:/config.json  bootstrap
docker rmi bootstrap

#Removing temp dir
rm -rf $DIR
