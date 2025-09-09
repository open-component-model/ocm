FROM --platform=$BUILDPLATFORM golang:1.25.1 AS builder
ARG COMMIT EFFECTIVE_VERSION GIT_TREE_STATE
ARG TARGETOS TARGETARCH

WORKDIR /go/src/ocm.software/ocm/
COPY go.* *.go VERSION ./
COPY api api
COPY cmds cmds
COPY hack/generate-docs hack/generate-docs
#COPY go/api api
RUN go get -d ./...
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
        go build -o /main -ldflags "-s -w \
	-X ocm.software/ocm/api/version.gitVersion=$EFFECTIVE_VERSION \
	-X ocm.software/ocm/api/version.gitTreeState=$GIT_TREE_STATE \
	-X ocm.software/ocm/api/version.gitCommit=$COMMIT \
	-X ocm.software/ocm/api/version.buildDate=$(date -u +%FT%T%z)" \
	./cmds/helminstaller

###################################################################################
FROM alpine:3.20.2

COPY --from=builder /main /toi/run
ENTRYPOINT [ "/toi/run" ]
