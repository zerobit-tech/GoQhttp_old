#!/usr/bin/env bash

APP_VERSION='v1.3.2'

output_name='QHttp_' 
rm ./bin/${output_name}${APP_VERSION}


echo 'Building..: '${output_name}${APP_VERSION}
go build  -ldflags="-X 'main.FeatureSet=ALL'  -X 'main.Version=${APP_VERSION}'" -o ./bin/${output_name}${APP_VERSION}  ./cmd/web



ls -al ./bin
#scp bin/${output_name}${APP_VERSION} sumit@154.49.243.18:~/qhttp/
docker build -t zerobittech/qhttp .
docker login
docker push zerobittech/qhttp
# copy db folder
#scp -r db/ sumit@154.49.243.18:~/qhttp/

# scp -r env/ sumit@154.49.243.18:~/qhttp/

# to stat the app
# nohup ./qhttp/QHttp_v1.3.2 --https=false &


#sudo docker run  -p 4091:4091 -e DOMAIN=qhttp.zerobit.tech -v /home/sumit/qhttp/lic:/app/lic -v /home/sumit/qhttp/db:/app/db -v /home/sumit/qhttp/env:/app/env --name=qhttp -d zerobittech/qhttp
