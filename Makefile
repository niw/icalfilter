NAME = icalfilter
TARGET = $(NAME)

.DEFAULT_GOAL := $(TARGET)

.PHONY: clean
clean:
	$(RM) -r libical
	$(RM) $(TARGET)

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
		-DCMAKE_DISABLE_FIND_PACKAGE_ICU=true \
		-DCMAKE_DISABLE_FIND_PACKAGE_BDB=true \
		./../../vendor/libical && \
	$(MAKE) install

# Build prerequisite other than libical should be handled by `go build`.
# Make the goal as a phony goal.
.PHONY: $(TARGET)
$(TARGET): libical/lib/libical.a libical/include/libical/ical.h
	go build -o "$@" ./cmd/icalfilter

.PHONY: docker
docker:
	docker build -t $(NAME) .

.PHONY: run
run: docker
	docker run --rm -i -t -p 3000:3000 -e PORT=3000 $(NAME)
