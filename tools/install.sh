#!/bin/bash
set -e

echo "installing zetup..."


default_release="0.0.1-alpha"

arch_info="$(uname -as)"
case $arch_info in
  *"x86_64"*) default_arch="amd64"
  * ) default_arch="i386"
esac

os_info="$(uname -as)"
case $os_info in
  *"Linux"*) default_os="linux"
  * ) default_os="darwin"
esac

ZETUP_OS=${ZETUP_OS:default_os}
ZETUP_ARCH=${ZETUP_ARCH:default_arch}
ZETUP_RELEASE=${ZETUP_RELEASE:default_release}


echo "$ZETUP_RELEASE, $ZETUP_ARCH, $ZETUP_OS"

url="https://github.com/zetup-sh/zetup/releases/download/$ZETUP_RELEASE/zetup-$ZETUP_OS-$ZETUP_ARCH"

echo the url is ...
echo $url