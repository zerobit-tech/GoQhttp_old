QSQL: SQL Client for IBM i DB2 written in Go

Raspberry pi build
env GOOS=linux GOARCH=arm go build -o goMockAPI_rpi  ./cmd/web


scp goMockAPI_rpi admin@loc:~


// linux batch
nohup ~.goMockAPI_rpi &


docker run --name gomockapi_lxpose --network=host -e ACCESS_TOKEN=xIooamze6Kch6AmKvsQDBo9CJU5WSKMGt8NYKsqs localxpose/localxpose:latest tunnel http --https-to 4041 --reserved-domain gomockapi.zerobit.tech --to localhost:4041



================================
If the actual site is running its own TLS, then you should tell LocalXpose to not terminate the TLS traffic and just forward it to your 443, so you can do it this way:

loclx tunnel http —to 80 —https-to 443

