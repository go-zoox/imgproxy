# Builder
# FROM --platform=$BUILDPLATFORM whatwewant/builder-go:v1.24-1 as builder

# https://blog.csdn.net/sun007700/article/details/120487881
FROM whatwewant/builder-go:v1.24-1 as builder


RUN apt update -y && apt install -y libvips

WORKDIR /build

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY . ./

ARG TARGETOS

ARG TARGETARCH

RUN CGO_ENABLED=1 \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  go build \
  -trimpath \
  -ldflags '-w -s' \
  -v -o imgproxy ./cmd/imgproxy

# Product
FROM whatwewant/alpine:v3.17-1

LABEL MAINTAINER="Zero<tobewhatwewant@gmail.com>"

LABEL org.opencontainers.image.source="https://github.com/go-zoox/imgproxy"

RUN apk add --no-cache vips

COPY --from=builder /build/imgproxy /bin

CMD imgproxy
