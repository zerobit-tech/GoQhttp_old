#!/usr/bin/env bash

output_name='QHttp_demo'
echo 'Building..: '$output_name
go build  -ldflags="-X 'main.FeatureSet=DEMO'" -o ./bin/$output_name ./cmd/web


output_name='QHttp_pub400' 
echo 'Building..: '$output_name
go build  -ldflags="-X 'main.FeatureSet=PUB400'" -o ./bin/$output_name ./cmd/web


output_name='QHttp' 
echo 'Building..: '$output_name
go build  -ldflags="-X 'main.FeatureSet=ALL'" -o ./bin/$output_name ./cmd/web
