#!/bin/bash


package="github.com/zetup-sh/zetup"

package_split=(${package//\// })
package_name=${package_split[-1]}

mkdir -p build
rm -rf build/*
cd build
platforms=("$1")
if [[ -z "$1" ]]; then
  echo "1 was not set"
  platforms=("linux/amd64" "linux/386" "windows/amd64" "windows/386" "darwin/amd64")
fi

for platform in "${platforms[@]}"
do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}
  output_name=$package_name'-'$GOOS'-'$GOARCH
  if [ $GOOS = "windows" ]; then
    output_name+='.exe'
  fi

  env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
  if [ $? -ne 0 ]; then
    echo 'An error has occurred! Aborting the script execution...'
    exit 1
  fi
  if [ "${ZETUP_NO_ZIP}" != "true" ] ; then
    if [ -x "$(command -v zip)" ] ; then
      zip  "$output_name.zip" "$output_name"
    elif [ -x "$(command -v 7z)" ] ; then
      7z a "$output_name.zip" "$output_name"
    fi
  fi
done
