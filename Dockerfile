FROM alpine:latest as builder

RUN mkdir -p /build
WORKDIR build

RUN apk add --no-cache \
    cmake \
    g++ \
    gcc \
    go \
    libc-dev \
    make \
    perl

COPY vendor/libical /build/vendor/libical

RUN mkdir -p libical/build \
    && cd libical/build \
    && cmake \
        -DCMAKE_BUILD_TYPE=Release \
        -DWITH_CXX_BINDINGS=false \
        -DICAL_ALLOW_EMPTY_PROPERTIES=true \
        -DSTATIC_ONLY=true \
        -DICAL_BUILD_DOCS=false \
        -DICAL_GLIB=false \
        -DCMAKE_INSTALL_PREFIX=`pwd`/.. \
        ./../../vendor/libical \
    && make install

COPY *.go /build

RUN go build -o shrinkical --ldflags '-extldflags "-static"'

FROM alpine:latest as runner
COPY --from=builder /build/shrinkical ./
CMD ./shrinkical
