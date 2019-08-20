#!/bin/bash

public_user_info=$(curl --url "https://gitlab.com/api/v4/users?username=$gl_username")
echo "$public_user_info"
gl_id=$(echo "$public_user_info" | yq -r ".[0].id")
echo "$gl_id"
curl \
  -u "$gl_username:$gl_password" \
  --url "https://gitlab.com/api/v4/users/$gl_id/keys"


# curl \
#   -u "zwhitchcox:$gl_pass"