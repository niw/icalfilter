FROM alpine:latest as builder

RUN apk add --no-cache \
    cmake \
    g++ \
    gcc \
    git \
    go \
    libc-dev \
    make \
    perl

RUN mkdir -p /gopath/src/github.com/niw/icalfilter

WORKDIR /gopath/src/github.com/niw/icalfilter
ENV GOPATH=/gopath

COPY vendor/libical ./vendor/libical/
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
        -DCMAKE_DISABLE_FIND_PACKAGE_ICU=true \
        ./../../vendor/libical \
    && make install

COPY *.go ./
COPY cmd/icalfilterd ./cmd/icalfilterd/
RUN go get --ldflags '-s -w -extldflags "-static"' github.com/niw/icalfilter/...


FROM alpine:latest

RUN apk add --no-cache \
  ca-certificates

COPY --from=builder /gopath/bin/icalfilterd ./

# Use `PORT` environment variable to set listening port, which is compatible with Heroku.
# Use 29s for timeout for Heroku H12 request timeout error.
CMD ./icalfilterd -addr 0.0.0.0 -port $PORT -timeout 29000
