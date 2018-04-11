FROM golang:1.10
WORKDIR /go/src/github.com/sevein/k8s-sysdig-adapter

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/sevein/k8s-sysdig-adapter/cmd/adapter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/adapter /bin/adapter
