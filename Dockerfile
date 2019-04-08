FROM golang:1.12 as builder
WORKDIR /go/src/github.com/draios/kubernetes-sysdig-metrics-apiserver
COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/draios/kubernetes-sysdig-metrics-apiserver/cmd/adapter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/adapter /bin/adapter
CMD ["/bin/adapter"]
