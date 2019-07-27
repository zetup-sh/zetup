#!/bin/bash

# GOAL: `init.sh` should be able to be run multiple times without negative effects
sudo echo have sudo privileges > /dev/null # add sudo privileges early

ZETUP_INSTALLATION_ID="zetup $(hostname) $USER $RANDOM"

# create directories
ZETUP_BACKUP_DIR="$HOME/.zetup/.bak"
mkdir -p $ZETUP_BACKUP_DIR
ZETUP_CONFIG_DIR="$HOME/.config/zetup"
mkdir -p $ZETUP_CONFIG_DIR

ensure_command_apt() {
  # makes sure command exists
  # pass list of command followed by packages where the commands are located
  # if the command is not found, it will install the package with apt
  # useage: ensure_command_apt [COMMAND] [PACKAGE] [COMMAND] [PACKAGE]...
  for (( i=1; i<=$#; i+=2)); do
    j=$((i+1))
    cmd="${!i}"
    pkg="${!j}"
    if [ ! -x "$(command -v $cmd)" ]
    then
      if [ -z $update_run ]
      then
        sudo apt update -qq
        update_run=true
      fi
      sudo apt install $pkg -qq -y
    fi
    echo "${!i}" "${!j}"
  done
  unset update_run
}

ensure_command_snap() {
  # makes sure command exists
  # pass list of command followed by packages where the commands are located
  # if the command is not found, it will install the package with snap
  # useage: ensure_command_snap [COMMAND] [PACKAGE] [COMMAND] [PACKAGE]...
  for (( i=1; i<=$#; i+=2)); do
    j=$((i+1))
    cmd="${!i}"
    pkg="${!j}"
    if [ ! -x "$(command -v $cmd)" ]
    then
      snap install $pkg -qq
    fi
  done
}

# TODO gitlab support
get_token() {
  # check for environment variable
  if [ -z $ZETUP_GITHUB_TOKEN ];
  then
    # if token info file doesn't exist, generate one
    TOKEN_INFO_FILE=$HOME/.config/zetup/github_personal_access_token_info.json
    if [ ! -f $TOKEN_INFO_FILE ]
    then
      #prompt for username and password
      echo No ZETUP_GITHUB_TOKEN detected, generating
      if [ -z $GITHUB_USERNAME ]
      then 
        printf "Github Username: " 
        read -r GITHUB_USERNAME
      else
        echo "Using Github Username \`$GITHUB_USERNAME\`"
      fi
      if [ -z $GITHUB_PASSWORD ]
      then 
        printf "Github Password: "
        read -rs GITHUB_PASSWORD
        echo
      else
        echo "Using \$GITHUB_PASSWORD"
      fi
      # dependencies to generate token
      ensure_command_snap jq jq curl curl

      # request generates token with all privileges TODO find privileges I actually need
      # keys I for sure need: "write:public_key"
      RESPONSE_FILE="/tmp/zetup-$RANDOM"
      data='{
      "note": "'"$ZETUP_INSTALLATION_ID"'",
      "scopes": [
      "repo",
      "admin:org",
      "admin:public_key",
      "admin:repo_hook",
      "gist",
      "notifications",
      "user",
      "delete_repo",
      "write:discussion",
      "admin:gpg_key"
      ]
    }'
  status=$(curl -s --request POST -w %{http_code} -o $RESPONSE_FILE \
    --url https://api.github.com/authorizations \
    -u "$GITHUB_USERNAME:$GITHUB_PASSWORD" \
    --header 'content-type: application/json' \
    --data "$data" \
  )
    if [[ "$status" =~ ^2[0-9]+$ ]]; # ensure successful response code
    then
      GITHUB_PERSONAL_ACCESS_TOKEN_INFO=$(cat $RESPONSE_FILE | jq -r '{id: .id, token: .token}')
      echo $GITHUB_PERSONAL_ACCESS_TOKEN_INFO > $TOKEN_INFO_FILE
    else
      cat $RESPONSE_FILE
      echo "There was an error adding the personal token to your account $status";
      exit 1
    fi
    fi
    ZETUP_GITHUB_TOKEN=$(cat $TOKEN_INFO_FILE | jq -r .token)
  else
    echo "using \$ZETUP_GITHUB_TOKEN"
    fi
  }

get_user_info() {
  USER_INFO_FILE="$ZETUP_CONFIG_DIR/user_info.json"
  if [ -f $USER_INFO_FILE ]
  then
    USER_INFO=$(cat $USER_INFO_FILE)
  else
    USER_INFO=$(curl -s -H "Authorization: token $ZETUP_GITHUB_TOKEN" https://api.github.com/user)
    echo $USER_INFO > $USER_INFO_FILE
  fi
  USERNAME=$(echo $USER_INFO | jq -r ".login")
  EMAIL=$(echo $USER_INFO | jq -r ".email")
  NAME=$(echo $USER_INFO | jq -r ".name")
}




# get user info and token
echo getting token
get_token
echo getting user info
get_user_info

# dependencies
echo installing dependencies
sudo apt-get update > /dev/null
sudo apt-get install -y  -qq \
  tmux \
  cmake \
  apt-transport-https \
  wget \
  ca-certificates \
  software-properties-common \
  snapd \
  git \
  xclip > /dev/null

# setup git
echo creating .gitconfig
git config --global user.name "$NAME"
git config --global user.email "$EMAIL"

# generate ssh key and add to github
echo adding ssh key to github
if [ ! -f $HOME/.ssh/id_rsa ]; then
  ssh-keygen -t rsa -b 4096 -C "$EMAIL" -f $HOME/.ssh/id_rsa -N ""
fi
eval $(ssh-agent -s) > /dev/null
ssh-add  $HOME/.ssh/id_rsa 2>/dev/null
RESPONSE_FILE="/tmp/zetup$RANDOM"
status=$(curl -s -u $USERNAME:$ZETUP_GITHUB_TOKEN -w %{http_code} -o $RESPONSE_FILE\
  --request POST  \
  --url https://api.github.com/user/keys \
  --data '{
  "title": "'"$ZETUP_INSTALLATION_ID"'",
  "key": "'"$(cat $HOME/.ssh/id_rsa.pub)"'"
}' \
)
if [[ "$status" =~ ^2[0-9]+$ ]]; # ensure successful response code
then
  echo ssh id added to github
else
  if ! [[ "$(cat $RESPONSE_FILE | jq -r '.errors[0].message')" == key\ is\ already* ]]
  then 
    cat $RESPONSE_FILE
    echo "There was a $status error adding id_rsa.pub to your account";
    exit 1
  fi
fi

# add gh/gl to known hosts
echo adding gh/gl to known hosts
ssh-keyscan -H github.com >> $HOME/.ssh/known_hosts 2>/dev/null
ssh-keyscan -H gitlab.com >> $HOME/.ssh/known_hosts 2>/dev/null

# autostar default config, easy to remember where it is
echo auto starring my own repository
if [[ $@ != "--no-star" ]];
then
  curl -s -u $USERNAME:$ZETUP_GITHUB_TOKEN \
    --request PUT \
    https://api.github.com/user/starred/zwhitchcox/zetup-config;
fi 

exit
# fork zetup if it doesn't exist
git ls-remote "git@github.com:$USERNAME/zetup-config.git" -q
if [ $? = 0 ]
then
  zetup_exists=true
else
  zetup_exists=false
fi




exit





#git clone "https://github.com/zwhitchcox/zetup.git" $HOME/zetup 
## uncomment the below line and comment the above line to use your own repo
##git clone "git@github.com/$USERNAME/zetup.git" $HOME/zetup 
#cd $HOME/dotfiles
#cp -r .bin ~/.bin
#git clone git@github.com:zwhitchcox/zetup.git $HOME/zetup
#cd $HOME/zetup
#mkdir ~/dev
#find . -maxdepth 1 -regextype posix-egrep -regex "\.\/\..*" ! -name .git -exec cp -t .. {} +
#sed -i "1s/^export username=$USERNAME/\n/" ~/.bashrc
#snap install yq
#source setup.sh
for i in $HOME/.zetup/cur/dotfiles/*;
do
  bn=$(basename $i) ;
  if [[ ! "$bn" = _* ]] ;
  then
    dotname="$HOME/.$bn"
    if [ -f "$dotname" ];
    then
      if [ ! -f "$dotname.zetup.bak" ]
      then
        mv "$dotname" "$dotname.zetup.bak"
      else
        rm "$dotname" # already have a backup of the original don't overwrite
      fi
    fi
    ln -s "$i" "$dotname"
  else
    dotname="$HOME/.$bn"
    if [ -f "$dotname" ];
    then
      rm "$dotname"
    fi
    ln -s "$i" "$dotname"
  fi
done
