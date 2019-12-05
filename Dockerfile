FROM golang:1.13.4-buster AS builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on

COPY . publish/

RUN cd publish && \
  go test ./... && \
  go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /usr/local/bin/publish main/main.go

FROM scratch AS release

COPY --from=builder /usr/local/bin/publish /usr/local/bin/publish

ENTRYPOINT ["/usr/local/bin/publish"]
