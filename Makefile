NAME = shrinkical
TARGET = $(NAME)

.PHONY: all
all: $(NAME)

.PHONY: clean
clean:
	rm -r libical
	rm $(TARGET)

libical/lib/libical.a: vendor/libical/CMakeLists.txt
	mkdir -p libical/build && \
	cd libical/build && \
	cmake \
		-DCMAKE_BUILD_TYPE=Debug \
		-DWITH_CXX_BINDINGS=false \
		-DICAL_ALLOW_EMPTY_PROPERTIES=true \
		-DSTATIC_ONLY=true \
		-DICAL_BUILD_DOCS=false \
		-DICAL_GLIB=false \
		-DCMAKE_INSTALL_PREFIX=`pwd`/.. \
		./../../vendor/libical && \
	make install

$(TARGET): libical/lib/libical.a libical/include/libical/ical.h main.go
	go build -o $@ --ldflags '-extldflags "-static"'

.PHONY: docker
docker:
	docker build -t $(NAME) .
