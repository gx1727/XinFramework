#!/bin/bash

OutputName="xin-server"
BuildPath="./cmd/server/main.go"
OutDir="./out"

# Create output directory
if [ ! -d "$OutDir" ]; then
    mkdir -p "$OutDir"
fi

echo "Building $OutputName..."

go build -ldflags="-s -w" -o "$OutDir/$OutputName" $BuildPath

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Output: $OutDir/$OutputName"

    FileSize=$(du -h "$OutDir/$OutputName" | cut -f1)
    echo "Size: $FileSize"

    # Copy config files to output directory
    echo "Copying configuration files..."
    if [ -d "./config" ]; then
        if [ ! -d "$OutDir/config" ]; then
            mkdir -p "$OutDir/config"
        fi
        cp -r ./config/* "$OutDir/config/"
        echo "Config files copied to $OutDir/config/"
    fi

    # Copy migrations if exists
    if [ -d "./migrations" ]; then
        if [ ! -d "$OutDir/migrations" ]; then
            mkdir -p "$OutDir/migrations"
        fi
        cp -r ./migrations/* "$OutDir/migrations/"
        echo "Migration files copied to $OutDir/migrations/"
    fi

    echo ""
    echo "Release package ready in '$OutDir' directory!"
else
    echo "Build failed!"
fi
