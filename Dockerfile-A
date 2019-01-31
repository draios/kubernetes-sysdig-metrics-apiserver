FROM golang:1.10
WORKDIR /go/src/github.com/dcberg/kubernetes-sysdig-metrics-apiserver

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/dcberg/kubernetes-sysdig-metrics-apiserver/cmd/adapter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY /go/bin/adapter /bin/adapter
CMD ["/bin/adapter"]