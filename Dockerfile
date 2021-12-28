FROM golang:1.17-bullseye

RUN useradd --create-home --home-dir /home/test test
ENV USER test

WORKDIR /opt

COPY go.mod ./
RUN go mod download

COPY main.go linkScanner.go ./
RUN go build -o /usr/bin/link-scanner

USER test
WORKDIR /home/test

ENTRYPOINT ["link-scanner"]

