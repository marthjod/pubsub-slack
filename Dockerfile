FROM golang as builder

ARG commit

COPY . /build

WORKDIR /build
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go get \
    honnef.co/go/tools/cmd/staticcheck \
    golang.org/x/lint \
    github.com/kisielk/errcheck
RUN make checks

RUN go build \
    -a \
    -ldflags "-X main.commit=${commit} \
              -extldflags \"-static\"" \
    -o /server .


FROM alpine

COPY --from=builder /server /
ENTRYPOINT ["/server"]
