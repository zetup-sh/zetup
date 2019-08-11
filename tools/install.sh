#!/bin/sh
set -e

echo "installing zetup..."


default_release="0.0.1-alpha"

if echo "$(uname -as)" | grep -q "x86_64";
then
  default_arch="amd64"
else
  default_arch="i386"
fi

if echo "$(uname -ms)" | grep -q "Linux";
then
  default_os="linux"
else
  default_os="darwin"
fi

ZETUP_OS=${ZETUP_OS:-$default_os}
ZETUP_ARCH=${ZETUP_ARCH:-$default_arch}
ZETUP_RELEASE=${ZETUP_RELEASE:-$default_release}


echo "$ZETUP_RELEASE, $ZETUP_ARCH, $ZETUP_OS"

curl -L -O "https://github.com/zetup-sh/zetup/releases/download/$ZETUP_RELEASE/zetup-$ZETUP_OS-$ZETUP_ARCH"