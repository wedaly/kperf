FROM golang:1.20 AS build-stage

WORKDIR /gomod
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /output

WORKDIR /kperf-build
RUN --mount=source=./,target=/kperf-build,rw make build && PREFIX=/output make install

FROM gcr.io/distroless/static-debian12:nonroot AS release-stage

WORKDIR /

COPY --from=build-stage /output/bin/kperf /kperf

USER nonroot:nonroot

ENTRYPOINT ["/kperf"]
