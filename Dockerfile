FROM golang:1-alpine as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o main github.com/codemicro/walrss/walrss

FROM alpine
COPY --from=builder /build/main /
WORKDIR /run

ENV WALRSS_DIR /run
CMD ["../main"]