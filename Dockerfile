FROM golang:1.11-alpine as build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk --no-cache add ca-certificates \
    && apk --no-cache add --virtual build-deps git

COPY ./*.go /go/src/github.com/ymyzk/dump1090_exporter/
WORKDIR /go/src/github.com/ymyzk/dump1090_exporter

RUN go get \
 && go build -o /bin/dump1090_exporter

FROM quay.io/prometheus/busybox:latest

COPY --from=build /bin/dump1090_exporter /bin/dump1090_exporter

EXPOSE 9190
ENTRYPOINT [ "/bin/dump1090_exporter" ]

