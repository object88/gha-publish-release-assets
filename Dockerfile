FROM golang:1.13.4-buster AS builder

COPY . src/publish/

RUN cd src/publish && \
  ./build.sh

FROM scratch AS release

COPY --from=builder /go/src/publish/bin/publish /usr/local/bin/publish

ENTRYPOINT ["/usr/local/bin/publish"]
