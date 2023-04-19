# GoQHttp
Mock APIs with ease 

HTTP: ==> go run ./cmd/web 
HTTPS: ==> go run ./cmd/web --https


openssl req  -new  -newkey rsa:2048  -nodes  -keyout goqhttp.key  -out goqhttp.csr


openssl  x509  -req  -days 365  -in goqhttp.csr  -signkey goqhttp.key  -out goqhttp.crt