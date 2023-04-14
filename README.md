# GoMockAPI
Mock APIs with ease 

HTTP: ==> go run ./cmd/web 
HTTPS: ==> go run ./cmd/web --https


openssl req  -new  -newkey rsa:2048  -nodes  -keyout gomockapi.key  -out gomockapi.csr


openssl  x509  -req  -days 365  -in gomockapi.csr  -signkey gomockapi.key  -out gomockapi.crt