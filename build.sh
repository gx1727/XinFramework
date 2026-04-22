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
            mkdir -p "$OutDir/apps/$appName"
            if [ -f "$appDir/config.yaml" ]; then
                cp "$appDir/config.yaml" "$OutDir/apps/$appName/"
            fi
            if [ -d "$appDir/migrations" ]; then
                cp -r "$appDir/migrations" "$OutDir/apps/$appName/"
            fi
        done
        echo "App config & migrations copied to $OutDir/apps/"
    fi

    echo ""
    echo "Release package ready in '$OutDir' directory!"
else
    echo "Build failed!"
fi
