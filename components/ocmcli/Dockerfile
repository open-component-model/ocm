ARG SUFFIX
ARG OCM_VERSION
ARG ALPINE_VERSION
FROM --platform=$BUILDPLATFORM alpine:${ALPINE_VERSION}
ARG SUFFIX
ARG OCM_VERSION
ARG ALPINE_VERSION

LABEL org.opencontainers.image.description Open Component Model Client ${OCM_VERSION} based on Alpine ${ALPINE_VERSION}
COPY gen/ocmcli/ocmcli.$SUFFIX /usr/bin/ocm
