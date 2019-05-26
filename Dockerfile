FROM golang:latest as builder

WORKDIR $GOPATH/src/github.com/Alexander-r/gosrvmon

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get -tags purego modernc.org/ql
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get github.com/lib/pq
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get golang.org/x/net/icmp
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get golang.org/x/net/ipv4
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go get golang.org/x/net/ipv6

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/bin/gosrvmon .
RUN cp config.json /etc/gosrvmon.json

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/bin/gosrvmon /usr/bin/gosrvmon
COPY --from=builder /etc/gosrvmon.json /config.json

EXPOSE 8000

ENTRYPOINT ["/usr/bin/gosrvmon"]

CMD []
