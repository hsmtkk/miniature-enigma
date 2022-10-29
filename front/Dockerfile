FROM golang:1.19 AS builder
WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY front/ /go/src/app/front/
COPY util/ /go/src/app/util/
COPY trace/ /go/src/app/trace/
WORKDIR /go/src/app/front
RUN CGO_ENABLED=0 go build -o front

FROM gcr.io/distroless/static-debian11 AS runtime
COPY --from=builder /go/src/app/front/front /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
