FROM golang:1.11 AS builder

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

RUN update-ca-certificates

WORKDIR /repos
RUN git clone https://github.com/travis-ci/packer-templates-mac

WORKDIR $GOPATH/src/github.com/travis-ci/imaged
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /imaged github.com/travis-ci/imaged/cmd/imaged

FROM hashicorp/packer:light
ADD https://github.com/jetbrains-infra/packer-builder-vsphere/releases/download/v2.0/packer-builder-vsphere-clone.linux /bin/packer-builder-vsphere-clone.linux
ADD https://github.com/jetbrains-infra/packer-builder-vsphere/releases/download/v2.0/packer-builder-vsphere-iso.linux /bin/packer-builder-vsphere-iso.linux
RUN chmod +x /bin/packer-builder-vsphere-clone.linux /bin/packer-builder-vsphere-iso.linux
COPY --from=builder /repos/packer-templates-mac /templates
COPY --from=builder /imaged .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./imaged"]
