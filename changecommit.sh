#!/bin/bash

git filter-branch -f --env-filter '

oldEmail="110393804+MarkSAO-F@users.noreply.github.com"
newName="markSAO-F"
newEmail="mark@sao.network"

if [ "$GIT_COMMITTER_EMAIL" = "$oldEmail" ]; then
 export GIT_COMMITTER_NAME="$newName"
 export GIT_COMMITTER_EMAIL="$newEmail"
fi

if [ "$GIT_AUTHOR_EMAIL" = "$oldEmail" ]; then
 export GIT_AUTHOR_NAME="$newName"
 export GIT_AUTHOR_EMAIL="$newEmail"
fi
' --tag-name-filter cat -- --branches --tags
