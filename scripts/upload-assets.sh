FILE="./build/mysmallfile"
echo $ZETUP_GITHUB_TOKEN
curl -H "Authorization: token $ZETUP_GITHUB_TOKEN" -H "Content-Type: $(file -b --mime-type $FILE)" --data-binary @$FILE "https://uploads.github.com/repos/zetup-sh/zetup/releases/19270146/assets?name=$(basename $FILE)"