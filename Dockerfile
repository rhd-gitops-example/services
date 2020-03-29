FROM golang:latest AS build
COPY . /go/build
WORKDIR /go/build

RUN go build ./cmd/services

FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf install git
WORKDIR /root/
COPY --from=build /go/build/services /usr/local/bin
ENTRYPOINT ["/usr/local/bin/services"]
