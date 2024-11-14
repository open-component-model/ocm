ARG GO_VERSION="1.23"
ARG ALPINE_VERSION="3.20"
ARG DISTROLESS_VERSION=debian12:nonroot@sha256:d71f4b239be2d412017b798a0a401c44c3049a3ca454838473a4c32ed076bfea

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

WORKDIR /src
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum *.go VERSION ./

ARG GO_PROXY="https://proxy.golang.org"
ENV GOPROXY=${GO_PROXY}
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
	export VERSION=$(go run api/version/generate/release_generate.go print-rc-version) && \
    export NOW=$(date -u +%FT%T%z) && \
    go build -trimpath -ldflags \
    "-s -w -X ocm.software/ocm/api/version.gitVersion=$VERSION -X ocm.software/ocm/api/version.buildDate=$NOW" \
    -o /bin/ocm ./cmds/ocm/main.go

FROM gcr.io/distroless/static-${DISTROLESS_VERSION}
# pass arg from initial build
ARG DISTROLESS_VERSION

COPY --from=build /bin/ocm /usr/local/bin/ocm

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model command line interface based on Distroless ${DISTROLESS_VERSION}"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static-${DISTROLESS_VERSION}"

ENTRYPOINT ["/usr/local/bin/ocm"]
CMD ["version"]
