#!/bin/sh
set -e

# get sudo privileges early
sudo printf ""
echo "Installing zetup..."

if echo "$(uname -as)" | grep -q "x86_64";
then
  default_arch="amd64"
else
  default_arch="386"
fi

if echo "$(uname -ms)" | grep -q "Linux";
then
  default_os="linux"
else
  default_os="darwin"
fi

default_release="0.0.1-alpha"
default_install_location="/usr/local/bin/zetup"

ZETUP_OS=${ZETUP_OS:-$default_os}
ZETUP_ARCH=${ZETUP_ARCH:-$default_arch}
ZETUP_RELEASE=${ZETUP_RELEASE:-$default_release}
INSTALL_LOCATION=${INSTALL_LOCATION:-$default_install_location}
filename="zetup-$ZETUP_OS-$ZETUP_ARCH"

url="https://github.com/zetup-sh/zetup/releases/download/$ZETUP_RELEASE/$filename.zip"

tempdir="/tmp/zetup"
mkdir -p "$tempdir"
templocation="/tmp/$filename.zip"

if [ -x "$(command -v curl)"]
then
  curl -fsSL "$url"  -o "$templocation"
else
  wget "$url" -O "$templocation"
fi

fixed_location="/tmp/zetup-fixed.zip"
unzip -o "$templocation" -d "$tempdir" > /dev/null 2>&1  || true

chmod +x "$tempdir/$filename"
mv "$tempdir/$filename" "$tempdir/zetup"
sudo mv "$tempdir/$filename" "$INSTALL_LOCATION"

echo "You have successfully installed zetup to $INSTALL_LOCATION!"