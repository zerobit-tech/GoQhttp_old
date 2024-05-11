FROM debian

ARG APP_VERSION="v1.3.2"
 
ENV PORT=4091
EXPOSE ${PORT}

ENV DOMAIN=0.0.0.0
ENV USELETSENCRYPT=N

# RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
WORKDIR /app
COPY ./bin/QHttp_${APP_VERSION} ./QHttp
COPY ./drivers/ibm-iaccess.deb ./ibm-iaccess.deb

RUN apt-get -y update
RUN apt-get -y install build-essential
RUN apt-get install -y libx11-dev
RUN apt update && \
    apt install -y -q --no-install-recommends unixodbc-dev \
    unixodbc \
    libpq-dev libodbc1 odbcinst odbcinst1debian2 && \
    dpkg -i ./ibm-iaccess.deb && \
    rm ./ibm-iaccess.deb && \
    chmod +x  ./QHttp

CMD [ "./QHttp","--https=false" ]

# docker build -t zerobittech/qhttp .
# docker run -p 4081:4081 -v /home/sumit/ideaprojects/GoQhttp/bin/lic:/app/lic           --name=qhttp zerobittech/qhttp



# docker run -p 4081:4081 -v /home/sumit/ideaprojects/GoQhttp/bin/lic:/app/lic  -v /etc/odbc.ini:/etc/odbc.ini     -d    --name=qhttp zerobittech/qhttp