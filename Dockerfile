FROM golang:1.14.2 as builder

WORKDIR /go/src/noticoco
COPY . /go/src/noticoco

RUN go build server.go


FROM ubuntu:latest

WORKDIR /root/
COPY --from=builder /go/src/noticoco/server .

ENV PORT=8000
EXPOSE ${PORT}

CMD ["./server"]

