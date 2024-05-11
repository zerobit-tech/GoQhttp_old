# QHTTP: Expose IBM i Stored Procedures and Programs as Web Services

  

QHTTP is a Golang application designed to seamlessly expose IBM i (AS/400) stored procedures and programs as web services. With QHTTP , you can easily integrate your existing IBM i applications with modern web applications or other systems that communicate via HTTP.

  

## Features

  

- **Effortless  Integration**: Connect your IBM i applications to the web without the need for complex middleware or extensive modifications to your existing codebase.

- **Secure  Communication**: QHTTP supports HTTPS, ensuring secure communication between your IBM i server and clients.

- **High  Performance**: Built with Golang's concurrency and efficiency in mind, qhttp delivers high performance for handling web service requests.

- **Customizable**:  Easily configure endpoints, request handling, and security settings to suit your specific requirements.

- **Logging  and Monitoring**: QHTTP provides comprehensive logging capabilities, allowing you to monitor and analyze web service usage and performance.



## Getting Started

  

### Prerequisites

  

Before installing and running qhttp, ensure you have the following prerequisites:

  

- IBM i server with stored procedures or programs you want to expose as web services.

- Go 1.20.5 or latest installed on your development machine


### Installation

```bash

git clone https://github.com/zerobit-tech/GoQhttp.git

```


### Run QHTTP



```bash

go run ./cmd/web

```

By default, qhttp will listen on port 4081. You can customize the port by modifying the `PORT` env var if needed.

Default user name `admin2@example.com`
Default password `adminpass`



### Security Considerations

  

It's important to ensure that your QHTTP server is adequately secured, especially if it's exposed to the internet. QHTTP comes with an self signed certificate.
We recommand to use a reverse proxy service like [caddy](https://caddyserver.com/). 


### Environment variables
```bash
PORT=4081
ALLOWEDORIGINS=https://*,http://*
 

# Maximum entries for graph
MAX_GRAPH_ENTRIES=1000

# Max log entried for one end point
MAX_LOG_ENTRIES_FOR_ONE_ENDPOINT=1000
 


# Rate limit: 0 ==> disable

REQUESTS_PER_HOUR_BY_IP=1000
REQUESTS_PER_HOUR_BY_USER=1000


# server credentials

# {SERVERNAME}_USER=someuser
# {SERVERNAME}_PASSWORD=somepassword

PING_SERVER_EVERY=20s



ALLOWHTMLTEMPLATES=Y
```

## Docker file

pending


## License

  

This project is licensed under the [MIT License](LICENSE).