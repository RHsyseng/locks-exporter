FROM registry.access.redhat.com/ubi8/go-toolset AS builder
WORKDIR /opt/app-root/src
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-micro AS run
COPY --from=builder /opt/app-root/src/bin/locks-exporter /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/locks-exporter"]
