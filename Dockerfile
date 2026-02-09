FROM --platform=$BUILDPLATFORM golang:1.25.7-alpine3.22 AS build

RUN apk add --no-cache make git

WORKDIR /src

COPY go.mod go.sum *.go VERSION ./

ARG GO_PROXY="https://proxy.golang.org"
ENV GOPROXY=${GO_PROXY}
RUN go mod download

COPY . .

ENV BUILD_FLAGS="-trimpath"

# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN make bin/ocm GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH}

FROM gcr.io/distroless/static-debian12:nonroot@sha256:e8a4044e0b4ae4257efa45fc026c0bc30ad320d43bd4c1a7d5271bd241e386d0

COPY --from=build /src/bin/ocm /usr/local/bin/ocm

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model command line interface based on Distroless"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static-debian12:nonroot"

ENTRYPOINT ["/usr/local/bin/ocm"]
CMD ["version"]
