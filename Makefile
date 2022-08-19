# build config
BUILD_DIR       ?= $(abspath build)
GET_GOARCH       = $(word 2,$(subst -, ,$1))
GET_GOOS         = $(word 1,$(subst -, ,$1))
GOBUILD         ?= $(shell go env GOOS)-$(shell go env GOARCH)
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
VETARGS?         =-all
GIT_COMMIT      ?= $(shell git describe --tags)
GIT_DIRTY       := $(if $(shell git status --porcelain),+CHANGES)
GO_LDFLAGS      := "-X main.Version=$(GIT_COMMIT)$(GIT_DIRTY)"

$(BUILD_DIR):
	mkdir -p $@

# Install requirements needed for the project to build
.PHONY: requirements
requirements:
	@go get

# Lazy alias for "go install" with custom GO_LDFLAGS enabled
.PHONY: install
install: requirements
	@go install

# Shortcut for `go build` with custom GO_LDFLAGS enabled
.PHONY: build
build: requirements
	$(MAKE) -j $(BINARIES)

BINARIES = $(addprefix $(BUILD_DIR)/hashi-helper-, $(GOBUILD))
$(BINARIES): $(BUILD_DIR)/hashi-helper-%: $(BUILD_DIR)
	@echo "=> building $@ ..."
	GOOS=$(call GET_GOOS,$*) GOARCH=$(call GET_GOARCH,$*) CGO_ENABLED=0 go build -o $@ -ldflags $(GO_LDFLAGS)

.PHONY: docker
docker: docker-build
	@echo "=> push Docker image ..."
	docker tag seatgeek/hashi-helper:$(GIT_COMMIT) seatgeek/hashi-helper:$(TAG)
	docker push seatgeek/hashi-helper:$(TAG)

.PHONY: docker
docker-build:
	@echo "=> build and push Docker image ..."
	docker build -f Dockerfile -t seatgeek/hashi-helper:$(GIT_COMMIT) .

.PHONY: dist
dist: install
	@echo "=> building ..."
	$(MAKE) -j $(BINARIES)

.PHONY: ci
ci: install
	@echo "=> Running CI"
	@./ci.sh
