FROM golang:1.24-alpine

WORKDIR /app

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

COPY go.mod ./
COPY go.sum ./

RUN go mod download

ENV CGO_ENABLED=0
ENV GOOS=linux

COPY stack.env /app/stack.env

RUN go build -o go-findmy ./cmd

ENTRYPOINT ["/app/go-findmy"]