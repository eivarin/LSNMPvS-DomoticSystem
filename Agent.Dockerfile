FROM golang:1.22

RUN mkdir /src
RUN mkdir /app
WORKDIR /src
ADD . /src
RUN go build -o /app/agent /src/cmd/Agent/Main.go
VOLUME [ "/app/config.yml" ]
CMD ["/app/agent", "/app/config.yml"]