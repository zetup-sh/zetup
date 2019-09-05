#!/bin/bash
GITHUB_PASS="$1"

echo $PATH
ls -al $HOME
cat ~/.bashrc
zetup cache set subpkgs ''
zetup id use gh/zwhitchcox:$GITHUB_PASS
zetup use github.com/zetup-sh/zetup-pkg