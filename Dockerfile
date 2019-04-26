FROM alpine:latest as cabundle

RUN apk add --no-cache ca-certificates

FROM golang as build

ADD ec2price.go .
RUN go get -d
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
        -a \
        -installsuffix netgo \
        -tags netgo \
        -ldflags '-w -extldflags "-static"' \
        -o ec2price ec2price.go

FROM scratch

COPY --from=cabundle /etc/ssl/certs /etc/ssl/certs
COPY --from=build    /go/ec2price  /ec2price

EXPOSE 8000

CMD ["/ec2price"]
