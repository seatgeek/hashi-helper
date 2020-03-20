# build config
BUILD_DIR 	?= $(abspath build)
GET_GOARCH 	= $(word 2,$(subst -, ,$1))
GET_GOOS 	= $(word 1,$(subst -, ,$1))
GOBUILD 	?= $(shell go env GOOS)-$(shell go env GOARCH)
GIT_COMMIT 	:= $(shell git describe --tags)
GIT_DIRTY 	:= $(if $(shell git status --porcelain),+CHANGES)
GO_LDFLAGS 	:= "-X main.Version=$(GIT_COMMIT)$(GIT_DIRTY)"

$(BUILD_DIR):
	mkdir -p $@

.PHONY: install
install:
	@echo "=> go mod download"
	@go mod download

.PHONY: build
build: install
	@echo "==> go install"
	@go install

.PHONY: ci
ci: install
	@echo "=> Running CI"
	@./ci.sh

BINARIES = $(addprefix $(BUILD_DIR)/hashi-helper-, $(GOBUILD))
$(BINARIES): $(BUILD_DIR)/hashi-helper-%: $(BUILD_DIR)
	@echo "=> building $@ ..."
	GOOS=$(call GET_GOOS,$*) GOARCH=$(call GET_GOARCH,$*) CGO_ENABLED=0 go build -o $@ -ldflags $(GO_LDFLAGS)

.PHONY: dist
dist: install
	@echo "=> building ..."
	$(MAKE) -j $(BINARIES)

.PHONY: docker
docker:
	@echo "=> build and push Docker image ..."
	docker build -f Dockerfile -t seatgeek/hashi-helper:$(COMMIT) .
	docker tag seatgeek/hashi-helper:$(COMMIT) seatgeek/hashi-helper:$(TAG)
	docker push seatgeek/hashi-helper:$(TAG)
