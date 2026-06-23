FROM golang:1.25.11-trixie

RUN useradd --create-home --home-dir /home/test test
ENV USER test

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download

COPY main.go client.go ./
RUN go build -o /usr/bin/link-scanner

USER test
WORKDIR /home/test

ENTRYPOINT ["link-scanner"]

