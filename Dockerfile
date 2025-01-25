FROM golang:1 as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build
# Go 1.18 introduced a compile-time dependency for Git unless `-buildvcs=false` is provided
RUN CGO_ENABLED=1 GOOS=linux go build -a -buildvcs=false -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o main github.com/codemicro/walrss/walrss

FROM alpine
COPY --from=builder /build/main /
WORKDIR /run

ENV WALRSS_DIR /run
CMD ["../main"]