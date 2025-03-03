#!/bin/bash

if [ ! -f ~/go/bin/griddriver ]; then
    echo "Error: griddriver not found. Please run build.sh first."
    exit 1
fi

sudo cp ~/go/bin/griddriver /usr/local/bin
echo Installed griddriver to /usr/local/bin
