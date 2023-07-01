FROM golang:1.20-bullseye
ENV DEBIAN_FRONTEND noninteractive
# RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
WORKDIR /app
COPY . ./

RUN apt-get update && \
    apt-get install -y -q --no-install-recommends dialog apt-utils && \
    apt-get install -y -q --no-install-recommends unixodbc-dev \
    unixodbc \
    libpq-dev && \
    go mod download && \
 go build -o build/gomokeapi ./cmd/web && \
 chmod +x  ./build/gomokeapi
 


EXPOSE 4040
CMD [ "./build/gomokeapi" ]