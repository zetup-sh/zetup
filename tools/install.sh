#!/bin/bash

# have to manually specify the release
# as of now, only have prelease, and I don't
# really feel like writing a json parser
# just for pre production
if [ ! -z "$ZETUP_RELEASE" ]
then
  echo "You must manually set the release for now :("
  exit 1
fi

# set architecture if not already set
if [ ! -z "$ZETUP_ARCH" ]
then
  if [[ "$(uname -as)" == *"x86_64"* ]]
  then ZETUP_ARCH="amd64"
  else ZETUP_ARCH="i386"
  fi
fi


# set os if not already set
if [ ! -z "$ZETUP_OS" ]
then
  if [[ "$(uname -ms)" == *"Linux"* ]]
  then ZETUP_OS="linux"
  else ZETUP_OS="darwin" # we know it's either mac or linux
  fi
fi



url="https://github.com/zetup-sh/zetup/releases/download/$ZETUP_RELEASE/zetup-$ZETUP_OS-$ZETUP_ARCH"

echo the url is ...
echo $url