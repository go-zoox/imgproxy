# Builder
# FROM --platform=$BUILDPLATFORM whatwewant/builder-go:v1.24-1 as builder

# https://blog.csdn.net/sun007700/article/details/120487881
FROM whatwewant/builder-go:v1.24-1 as builder

RUN apt update -y && apt install -y pkg-config libvips-dev && ldconfig

# # Installs libvips + required libraries
# #  reference: https://github.com/h2non/imaginary/blob/master/Dockerfile
# ARG LIBVIPS_VERSION=8.16.1
# #
# RUN DEBIAN_FRONTEND=noninteractive \
#   apt-get update && \
#   apt-get install --no-install-recommends -y \
#   ca-certificates \
#   automake build-essential curl \
#   gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg62-turbo-dev libpng-dev \
#   libwebp-dev libtiff5-dev libgif-dev libexif-dev libxml2-dev libpoppler-glib-dev \
#   swig libmagickwand-dev libpango1.0-dev libmatio-dev libopenslide-dev libcfitsio-dev \
#   libgsf-1-dev fftw3-dev liborc-0.4-dev librsvg2-dev libimagequant-dev libheif-dev && \
#   cd /tmp && \
#   curl -fsSLO https://github.com/libvips/libvips/releases/download/v${LIBVIPS_VERSION}/vips-${LIBVIPS_VERSION}.tar.gz && \
#   tar zvxf vips-${LIBVIPS_VERSION}.tar.gz && \
#   cd /tmp/vips-${LIBVIPS_VERSION} && \
# 	CFLAGS="-g -O3" CXXFLAGS="-D_GLIBCXX_USE_CXX11_ABI=0 -g -O3" \
#     ./configure \
#     --disable-debug \
#     --disable-dependency-tracking \
#     --disable-introspection \
#     --disable-static \
#     --enable-gtk-doc-html=no \
#     --enable-gtk-doc=no \
#     --enable-pyvips8=no && \
#   make && \
#   make install && \
#   ldconfig

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
# FROM whatwewant/alpine:v3.17-1
FROM whatwewant/ubuntu:v22.04-1

LABEL MAINTAINER="Zero<tobewhatwewant@gmail.com>"

LABEL org.opencontainers.image.source="https://github.com/go-zoox/imgproxy"

RUN apt update -y && apt install -y pkg-config libvips-dev && ldconfig

# RUN apk add --no-cache vips

COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /build/imgproxy /bin

CMD imgproxy
