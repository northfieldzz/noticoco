FROM golang:1.14.2

WORKDIR /go/src/noticoco
COPY . /go/src/noticoco

RUN go mod download
ENV PORT=8000
EXPOSE ${PORT}
CMD ["go", "run", "server.go"]

