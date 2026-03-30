#!/bin/bash
set -e
echo "PWD: $(pwd)" >> /tmp/sign-debug.log
echo "Artifact: $1" >> /tmp/sign-debug.log
echo "Signature: $2" >> /tmp/sign-debug.log
ls -la "$1" >> /tmp/sign-debug.log 2>&1
gpg --batch --yes --output "$2" --detach-sign "$1"
gpg --verify "$2" "$1" >> /tmp/sign-debug.log 2>&1
