FROM --platform=$BUILDPLATFORM golang:1.23-alpine3.20 AS build

RUN apk add --no-cache make git

WORKDIR /src

COPY go.mod go.sum *.go VERSION ./

ARG GO_PROXY="https://proxy.golang.org"
ENV GOPROXY=${GO_PROXY}
RUN go mod download

COPY . .

ENV BUILD_FLAGS="-trimpath"

RUN make bin/ocm

FROM gcr.io/distroless/static-debian12:nonroot@sha256:d71f4b239be2d412017b798a0a401c44c3049a3ca454838473a4c32ed076bfea

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
