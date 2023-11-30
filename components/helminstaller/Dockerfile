FROM --platform=$BUILDPLATFORM golang:1.21 AS builder
ARG COMMIT EFFECTIVE_VERSION GIT_TREE_STATE
ARG TARGETOS TARGETARCH

WORKDIR /go/src/github.com/open-component-model/ocm/
COPY go.* ./
COPY pkg pkg
COPY cmds cmds
COPY hack/generate-docs hack/generate-docs
#COPY go/pkg pkg
RUN --mount=type=cache,target=/root/.cache/go-build go get -d ./...
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
        go build -o /main -ldflags "-s -w \
	-X github.com/open-component-model/ocm/pkg/version.gitVersion=$EFFECTIVE_VERSION \
	-X github.com/open-component-model/ocm/pkg/version.gitTreeState=$GIT_TREE_STATE \
	-X github.com/open-component-model/ocm/pkg/version.gitCommit=$COMMIT \
	-X github.com/open-component-model/ocm/pkg/version.buildDate=$(date -u +%FT%T%z)" \
	./cmds/helminstaller

###################################################################################
FROM alpine

COPY --from=builder /main /toi/run
ENTRYPOINT [ "/toi/run" ]
