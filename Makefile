include scripts/dqlite/Makefile

JUJU_GOMOD_MODE ?= readonly
LINK_FLAGS = "-s -w -extldflags '-static' $(link_flags_version)"
CGO_LINK_FLAGS = "-s -w -linkmode 'external' -extldflags '-static' $(link_flags_version)"
BUILD_TAGS ?= "libsqlite3 dqlite"

define run_cgo_install
	@echo "Installing ${PACKAGE}"
	@env PATH="${MUSL_BIN_PATH}:${PATH}" \
		CC="musl-gcc" \
		CGO_CFLAGS="-I${DQLITE_EXTRACTED_DEPS_ARCHIVE_PATH}/include" \
		CGO_LDFLAGS="-L${DQLITE_EXTRACTED_DEPS_ARCHIVE_PATH} -luv -lraft -ldqlite -llz4 -lsqlite3" \
		CGO_LDFLAGS_ALLOW="(-Wl,-wrap,pthread_create)|(-Wl,-z,now)" \
		LD_LIBRARY_PATH="${DQLITE_EXTRACTED_DEPS_ARCHIVE_PATH}" \
		CGO_ENABLED=1 \
		go install \
			-mod=$(JUJU_GOMOD_MODE) \
			-tags=$(BUILD_TAGS) \
			-ldflags ${CGO_LINK_FLAGS} \
			-v ${PACKAGE}
endef

build: PACKAGE = github.com/SimonRichardson/dqlite-bug-reproducer
build: musl-install-if-missing dqlite-install-if-missing
	${run_cgo_install}