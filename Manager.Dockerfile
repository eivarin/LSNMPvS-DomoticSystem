FROM golang:1.22

RUN mkdir /src
RUN mkdir /app
WORKDIR /src
ADD . /src
RUN go build -o /app/manager /src/cmd/Manager/Main.go
VOLUME [ "/app/config.yml" ]
CMD ["/app/manager", "/app/config.yml"]