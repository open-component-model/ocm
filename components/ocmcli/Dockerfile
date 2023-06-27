ARG SUFFIX
ARG OCM_VERSION
FROM --platform=$BUILDPLATFORM alpine:latest
ARG SUFFIX
ARG OCM_VERSION

LABEL org.opencontainers.image.description Open Component Model Client ${OCM_VERSION} based on Alpine alpine:latest
COPY gen/ocmcli/ocmcli.$SUFFIX /usr/bin/ocm
