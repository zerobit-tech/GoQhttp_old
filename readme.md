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



## Docker file




## License

  

This project is licensed under the [MIT License](LICENSE).