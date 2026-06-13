#!/bin/bash

VERSION=$(cat VERSION)

echo "Building version $VERSION..."

docker build -t habit-tracker-bot:$VERSION .
docker build -f cmd/migration/Dockerfile -t habit-tracker-migration:$VERSION .

echo "Done! Images built:"
echo "  habit-tracker-bot:$VERSION"
echo "  habit-tracker-migration:$VERSION"