# This Dockerfile is used by `make` via the Makefile

ARG DISTROLESS_VERSION=debian12:nonroot@sha256:d71f4b239be2d412017b798a0a401c44c3049a3ca454838473a4c32ed076bfea
FROM --platform=$BUILDPLATFORM gcr.io/distroless/static-${DISTROLESS_VERSION}
# pass arg from initial build
ARG DISTROLESS_VERSION

ARG SUFFIX
ARG OCM_VERSION

COPY gen/ocmcli/ocmcli.$SUFFIX /usr/local/bin/ocm

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model command line interface based on Distroless ${DISTROLESS_VERSION}"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.version="${OCM_VERSION}"
LABEL org.opencontainers.image.revision="${OCM_VERSION}"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static-${DISTROLESS_VERSION}"

ENTRYPOINT ["/usr/local/bin/ocm"]
CMD ["version"]
