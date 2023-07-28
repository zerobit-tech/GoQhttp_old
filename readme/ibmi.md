https://go.dev/doc/install/source

https://github.com/skytap/rclone/blob/master/Rclone%20on%20Power.pdf

=================================================================================


CGO_ENABLED=0 GOOS=aix GOARCH=ppc64 go build -o ibmie .

 go build -compiler gccgo -buildmode=c-archive -o i2  -gccgoflags="-lgo -maix64"  .


CGO_ENABLED=0 GOOS=aix GOARCH=ppc64 ~/go_build/go-linux-amd64-bootstrap/bin/go build -o ./bin/ibmie ./cmd/web


scp -P2222 ./ibmie sumitg@pub400.com:ibmie

ssh -p2222 sumitg@pub400.com

