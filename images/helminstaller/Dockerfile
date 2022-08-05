FROM golang:1.17 AS builder

WORKDIR /go/src/github.com/open-component-model/ocm/
COPY go.* ./
COPY pkg pkg
COPY cmds cmds
#COPY go/pkg pkg
RUN go get -d ./...
RUN CGO_ENABLED=0 go build -o /main ./cmds/helminstaller

FROM alpine

COPY --from=builder /main /ocm/run
ENTRYPOINT [ "/ocm/run" ]
