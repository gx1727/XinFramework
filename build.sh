#!/bin/bash

OutputName="xin"
BuildPath="./cmd/xin"
OutDir="./out"

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

    echo "Copying configuration files..."
    if [ -d "./config" ]; then
        mkdir -p "$OutDir/config"
        cp -r ./config/* "$OutDir/config/"
        echo "Config files copied to $OutDir/config/"
    fi

    echo "Copying migration files..."
    if [ -d "./migrations" ]; then
        cp -r ./migrations "$OutDir/migrations"
        echo "Migration files copied to $OutDir/migrations/"
    fi

    echo ""
    echo "Release package ready in '$OutDir' directory!"
else
    echo "Build failed!"
fi
