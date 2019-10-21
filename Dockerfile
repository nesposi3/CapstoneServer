FROM golang:latest as build

WORKDIR /go/src/server
COPY . .



RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

FROM alpine:latest

RUN apk --update add ca-certificates

WORKDIR /root/
COPY --from=build /go/src/server/server ./

CMD ["./server"]
