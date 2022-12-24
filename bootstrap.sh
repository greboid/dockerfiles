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

echo "Making temp dir"
DIR=$(mktemp -d)

#Downloading latest version of apk.static and SHA
GL=https://gitlab.alpinelinux.org/api/v4
GLBASE=https://gitlab.alpinelinux.org/api/v4/projects/alpine%2Fapk-tools
VERCOMMAND="curl -qSs $GLBASE/repository/tags | jq -r '.[0].name'"
VER=$(eval "$VERCOMMAND")
IDCOMMAND="curl -qSs https://gitlab.alpinelinux.org/api/v4/projects/alpine%2Fapk-tools/packages | jq -r '.[] | select(.name==\"$VER\") | select(.version==\"x86_64\").id'"
ID=$(eval "$IDCOMMAND")
INFOCOMMAND="curl -qSs $GLBASE/packages/$ID/package_files | jq -r '.[0] | .file_sha256'"
SHA=$(eval "$INFOCOMMAND")
curl -qSs -o $DIR/apk.static https://gitlab.alpinelinux.org/alpine/apk-tools/-/package_files/$ID/download
FILESHA=$(sha256sum $DIR/apk.static | awk '{print $1}')
if [ "$FILESHA" != "$SHA" ]; then
  echo "Downloaded apk.static SHA256 doesn't match, aborting."
  exit 2
fi
chmod +x $DIR/apk.static

echo "https://mirrors.melbourne.co.uk/alpine/latest-stable/main" > $DIR/repositories

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

/apk.static -X https://mirrors.melbourne.co.uk/alpine/latest-stable/main --keys-dir ../etc/apk/keys --root /alpine --initdb add alpine-baselayout alpine-keys apk-tools busybox libc-utils
rm -Rf /alpine/var/cache/apk/*
rm -rf /home /media/cdrom /media/floppy /media/usb /mnt /srv /usr/local/bin /usr/local/lib /usr/local/share
cp /repositories /alpine/etc/apk/repositories

buildah bud -t $1 -f /Dockerfile
buildah push --authfile /config.json $1
EOF
chmod +x $DIR/run.sh

#This dockerfile is used in the container to create the base image
cat << EOF > $DIR/Dockerfile2
FROM scratch
COPY /alpine /
CMD ["/bin/sh"]
EOF

git clone --depth 1 --filter=blob:none --sparse https://gitlab.alpinelinux.org/alpine/aports.git $DIR/aports;
git --git-dir $DIR/aports/.git --work-tree $DIR/aports sparse-checkout set "main/alpine-keys"

#This dockerfile bootstraps an alpine container so it can build the base image
cat << EOF > $DIR/Dockerfile
FROM scratch
COPY apk.static /apk.static
COPY repositories /repositories
COPY run.sh /run.sh
COPY Dockerfile2 /Dockerfile
COPY aports/main/alpine-keys/ /keys/
RUN ["/apk.static", "-X", "http://mirrors.melbourne.co.uk/alpine/latest-stable/main", "--keys-dir", "keys", "-p", "/", "--initdb", "add", "alpine-baselayout", "alpine-keys", "apk-tools", "busybox"]
CMD ["/run.sh"]
EOF

echo "Building bootstrap"
docker build -t bootstrap -f $DIR/Dockerfile $DIR

echo "Running bootstrap"
docker run --rm --privileged -v ~/.docker/config.json:/config.json  bootstrap
docker rmi bootstrap

#Removing temp dir
rm -rf $DIR
