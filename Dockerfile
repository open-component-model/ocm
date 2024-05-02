ARG GO_VERSION="1.22"
ARG ALPINE_VERSION="3.19"

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

WORKDIR /src
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum *.go VERSION ./

ARG GO_PROXY="https://proxy.golang.org"
ENV GOPROXY=${GO_PROXY}
RUN --mount=type=cache,target=/root/.cache/go-build go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
	export VERSION=$(go run pkg/version/generate/release_generate.go print-rc-version) && \
    export NOW=$(date -u +%FT%T%z) && \
    go build -trimpath -ldflags \
    "-s -w -X github.com/open-component-model/ocm/pkg/version.gitVersion=$VERSION -X github.com/open-component-model/ocm/pkg/version.buildDate=$NOW" \
    -o /bin/ocm ./cmds/ocm/main.go

FROM alpine:${ALPINE_VERSION}

# Create group and user
ARG UID=1000
ARG GID=1000
RUN addgroup -g "${GID}" ocmGroup && adduser -u "${UID}" ocmUser -G ocmGroup -D

COPY --from=build /bin/ocm /bin/ocm
COPY --chmod=0755 components/ocmcli/ocm.sh /bin/ocm.sh

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model command line interface based on Alpine ${ALPINE_VERSION}"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="alpine:${ALPINE_VERSION}"

USER ocmUser
ENTRYPOINT ["/bin/ocm.sh"]
CMD ["/bin/ocm"]