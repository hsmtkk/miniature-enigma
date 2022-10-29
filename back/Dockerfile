FROM golang:1.19 AS builder
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY back/ /go/src/app/back/
COPY openweather/ /go/src/app/openweather/
COPY util/ /go/src/app/util/
COPY trace/ /go/src/app/trace/
WORKDIR /go/src/app/back
RUN CGO_ENABLED=0 go build -o back

FROM gcr.io/distroless/static-debian11 AS runtime
COPY --from=builder /go/src/app/back/back /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
