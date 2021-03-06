# Build imaged in a separate container
FROM golang:1.11 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN update-ca-certificates

WORKDIR $GOPATH/src/github.com/travis-ci/imaged
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /imaged github.com/travis-ci/imaged/cmd/imaged

# Pull in the vsphere-images binary from its Docker image
FROM travisci/vsphere-images AS vsphere-images

# Use the official Packer image as our base
FROM hashicorp/packer:light

# Download the third-party builders we use
ADD https://github.com/jetbrains-infra/packer-builder-vsphere/releases/download/v2.0/packer-builder-vsphere-clone.linux /bin/packer-builder-vsphere-clone.linux
ADD https://github.com/jetbrains-infra/packer-builder-vsphere/releases/download/v2.0/packer-builder-vsphere-iso.linux /bin/packer-builder-vsphere-iso.linux
RUN chmod +x /bin/packer-builder-vsphere-clone.linux /bin/packer-builder-vsphere-iso.linux

# Copy things from the other stages
COPY --from=vsphere-images /bin/vsphere-images /bin/vsphere-images
COPY --from=builder /imaged .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["./imaged"]
