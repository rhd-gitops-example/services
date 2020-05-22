FROM ubuntu AS oc-base

# Install curl
RUN apt-get update 
RUN apt-get install -y curl

# Install oc
RUN curl -O https://mirror.openshift.com/pub/openshift-v4/clients/oc/4.3/linux/oc.tar.gz

RUN tar xvf oc.tar.gz
RUN chmod +x ./oc
RUN cp ./oc /usr/local/bin/

# Build services binary, rename and add to path
FROM golang:latest AS build
COPY . /go/build
WORKDIR /go/build

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/services

FROM registry.access.redhat.com/ubi8/ubi-minimal
RUN microdnf install git
WORKDIR /root/
COPY --from=build /go/build/services /usr/local/bin/oc-services
COPY --from=oc-base /usr/local/bin/oc /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/oc"]
