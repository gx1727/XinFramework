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
        if [ ! -d "$OutDir/config" ]; then
            mkdir -p "$OutDir/config"
        fi
        cp -r ./config/* "$OutDir/config/"
        echo "Config files copied to $OutDir/config/"
    fi

    if [ -d "./migrations" ]; then
        if [ ! -d "$OutDir/migrations" ]; then
            mkdir -p "$OutDir/migrations"
        fi
        cp -r ./migrations/* "$OutDir/migrations/"
        echo "Migration files copied to $OutDir/migrations/"
    fi

    if [ -d "./apps" ]; then
        for appDir in ./apps/*/; do
            appName=$(basename "$appDir")
            if [ -d "$appDir/migrations" ]; then
                mkdir -p "$OutDir/migrations/$appName"
                cp -r "$appDir/migrations/"* "$OutDir/migrations/$appName/"
            fi
            if [ -f "$appDir/config.yaml" ]; then
                mkdir -p "$OutDir/config/$appName"
                cp "$appDir/config.yaml" "$OutDir/config/$appName/"
            fi
        done
        echo "App files copied"
    fi

    echo ""
    echo "Release package ready in '$OutDir' directory!"
else
    echo "Build failed!"
fi
