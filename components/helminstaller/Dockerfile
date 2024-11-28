FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
ARG COMMIT EFFECTIVE_VERSION GIT_TREE_STATE
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/ocm.software/ocm/
COPY go.* *.go VERSION ./
COPY api api
COPY cmds cmds
COPY hack/generate-docs hack/generate-docs
#COPY go/api api
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
        go build -o /main -ldflags "-s -w \
	-X ocm.software/ocm/api/version.gitVersion=$EFFECTIVE_VERSION \
	-X ocm.software/ocm/api/version.gitTreeState=$GIT_TREE_STATE \
	-X ocm.software/ocm/api/version.gitCommit=$COMMIT \
	-X ocm.software/ocm/api/version.buildDate=$(date -u +%FT%T%z)" \
	./cmds/helminstaller

###################################################################################
FROM gcr.io/distroless/static-debian12:nonroot@sha256:d71f4b239be2d412017b798a0a401c44c3049a3ca454838473a4c32ed076bfea

# https://github.com/opencontainers/image-spec/blob/main/annotations.md#pre-defined-annotation-keys
LABEL org.opencontainers.image.description="Open Component Model Helm Installer based on Distroless"
LABEL org.opencontainers.image.vendor="SAP SE"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.url="https://ocm.software/"
LABEL org.opencontainers.image.source="https://github.com/open-component-model/ocm"
LABEL org.opencontainers.image.title="ocm"
LABEL org.opencontainers.image.documentation="https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm.md"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static-debian12:nonroot"

COPY --from=builder /main /toi/run
ENTRYPOINT [ "/toi/run" ]
