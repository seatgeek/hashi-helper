# build config
BUILD_DIR 	?= $(abspath build)
GET_GOARCH 	= $(word 2,$(subst -, ,$1))
GET_GOOS 	= $(word 1,$(subst -, ,$1))
GOBUILD 	?= $(shell go env GOOS)-$(shell go env GOARCH)

$(BUILD_DIR):
	mkdir -p $@

.PHONY: install
install:
ifeq (, $(shell which dep))
	@echo "=> installing 'dep'"
	go get github.com/golang/dep/cmd/dep
endif
	@echo "=> installing dependencies"
	@dep ensure -v -vendor-only

.PHONY: build
build: install
	go install

.PHONY: ci
ci: install
	@echo "=> Running CI"
	@./ci.sh

BINARIES = $(addprefix $(BUILD_DIR)/hashi-helper-, $(GOBUILD))
$(BINARIES): $(BUILD_DIR)/hashi-helper-%: $(BUILD_DIR)
	@echo "=> building $@ ..."
	GOOS=$(call GET_GOOS,$*) GOARCH=$(call GET_GOARCH,$*) CGO_ENABLED=0 go build -o $@

.PHONY: dist
dist: install fmt vet
	@echo "=> building ..."
	$(MAKE) -j $(BINARIES)

.PHONY: docker
docker:
	@echo "=> build and push Docker image ..."
	@docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)
	docker build -f Dockerfile -t seatgeek/hashi-helper:$(COMMIT) .
	docker tag seatgeek/hashi-helper:$(COMMIT) seatgeek/hashi-helper:$(TAG)
	docker push seatgeek/hashi-helper:$(TAG)
