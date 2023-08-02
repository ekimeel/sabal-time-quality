#!/bin/bash

# Extract name from build.properties
MOD_NAME=$(grep "^name" build.properties | cut -d'=' -f2)

# Replace / with _ in the module name
FILE_NAME=$(echo ${MOD_NAME} | tr '/' '_')

# get the latest git tag
GIT_TAG=$(git describe --tags `git rev-list --tags --max-count=1`)

# build the project
go build -gcflags="all=-N -l" -o ${FILE_NAME}.so -buildmode=plugin

# rename the file and append the git tag
mv ${FILE_NAME}.so ${FILE_NAME}_${GIT_TAG}.so
#git remote add origin https://github.com/ekimeel/sabal-stats.git
 #git branch -M main
 #git push -u origin main