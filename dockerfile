FROM debian:latest
 
ENV PORT=4081
EXPOSE ${PORT}

ENV DOMAIN=0.0.0.0
ENV USELETSENCRYPT=N

# RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
WORKDIR /app
COPY ./bin/QHttp ./QHttp
COPY ./drivers/ibm-iaccess.deb ./ibm-iaccess.deb

RUN apt update && \
    apt install -y -q --no-install-recommends unixodbc-dev \
    unixodbc \
    libpq-dev libodbc1 odbcinst odbcinst1debian2 && \
    dpkg -i ./ibm-iaccess.deb && \
    rm ./ibm-iaccess.deb && \
    chmod +x  ./QHttp

CMD [ "./QHttp" ]

# docker build -t onlysumitg/qhttp .
# docker run -p 4081:4081 -v /home/sumit/ideaprojects/GoQhttp/bin/lic:/app/lic           --name=qhttp onlysumitg/qhttp



# docker run -p 4081:4081 -v /home/sumit/ideaprojects/GoQhttp/bin/lic:/app/lic  -v /etc/odbc.ini:/etc/odbc.ini     -d    --name=qhttp onlysumitg/qhttp