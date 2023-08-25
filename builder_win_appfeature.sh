#!/usr/bin/env bash

APP_VERSION='v1.3.0'

GOOS='windows'
GOARCH='amd64'

output_name='QHttp_win_demo'
echo 'Building..: '$output_name
env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=DEMO' -X 'main.Version=${APP_VERSION}'" -o ./bin/${output_name}${APP_VERSION}.exe ./cmd/web


# output_name='QHttp_win_pub400' 
# echo 'Building..: '$output_name
# env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=PUB400'" -o ./bin/$output_name ./cmd/web


output_name='QHttp_win' 
echo 'Building..: '$output_name
env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=ALL' -X 'main.Version=${APP_VERSION}'" -o ./bin/${output_name}${APP_VERSION}.exe ./cmd/web
