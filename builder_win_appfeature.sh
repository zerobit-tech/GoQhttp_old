#!/usr/bin/env bash
GOOS='windows'
GOARCH='amd64'

output_name='QHttp_win_demo'
echo 'Building..: '$output_name
env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=DEMO'" -o ./bin/$output_name.exe ./cmd/web


# output_name='QHttp_win_pub400' 
# echo 'Building..: '$output_name
# env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=PUB400'" -o ./bin/$output_name ./cmd/web


output_name='QHttp_win' 
echo 'Building..: '$output_name
env GOOS=$GOOS GOARCH=$GOARCH go build  -ldflags="-X 'main.FeatureSet=ALL'" -o ./bin/$output_name.exe ./cmd/web
