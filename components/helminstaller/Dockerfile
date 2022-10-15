FROM golang:1.17 AS builder
ARG COMMIT
ARG EFFECTIVE_VERSION
ARG GIT_TREE_STATE

WORKDIR /go/src/github.com/open-component-model/ocm/
COPY go.* ./
COPY pkg pkg
COPY cmds cmds
#COPY go/pkg pkg
RUN --mount=type=cache,target=/root/.cache/go-build go get -d ./...
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o /main -ldflags "-s -w \
	-X github.com/open-component-model/ocm/pkg/version.gitVersion=$EFFECTIVE_VERSION \
	-X github.com/open-component-model/ocm/pkg/version.gitTreeState=$GIT_TREE_STATE \
	-X github.com/open-component-model/ocm/pkg/version.gitCommit=$COMMIT \
	-X github.com/open-component-model/ocm/pkg/version.buildDate=$(date --rfc-3339=seconds | sed 's/ /T/')" \
	./cmds/helminstaller

FROM alpine

COPY --from=builder /main /toi/run
ENTRYPOINT [ "/toi/run" ]
