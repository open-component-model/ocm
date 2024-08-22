# This Dockerfile is used by `make` via the Makefile

ARG ALPINE_VERSION="3.20.2"
FROM --platform=$BUILDPLATFORM alpine:${ALPINE_VERSION}

# Create group and user
ARG UID=1000
ARG GID=1000
RUN addgroup -g "${GID}" ocmGroup && adduser -u "${UID}" ocmUser -G ocmGroup -D

ARG SUFFIX
ARG OCM_VERSION

COPY gen/ocmcli/ocmcli.$SUFFIX /bin/ocm
COPY --chmod=0755 components/ocmcli/ocm.sh /bin/ocm.sh

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model command line interface based on Alpine ${ALPINE_VERSION}"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.version="${OCM_VERSION}"
LABEL org.opencontainers.image.revision="${OCM_VERSION}"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="alpine:${ALPINE_VERSION}"

USER ocmUser
ENTRYPOINT ["/bin/ocm.sh"]
CMD ["/bin/ocm"]
