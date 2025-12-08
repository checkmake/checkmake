FROM golang:1.25 as builder

ARG BUILDER_NAME
ARG BUILDER_EMAIL

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
COPY . /go/src/github.com/checkmake/checkmake

WORKDIR /go/src/github.com/checkmake/checkmake
RUN make binaries
RUN make test

FROM alpine:3.23
RUN apk add --no-cache make
USER nobody

COPY --from=builder /go/src/github.com/checkmake/checkmake/checkmake /
ENTRYPOINT ["./checkmake", "/Makefile"]
