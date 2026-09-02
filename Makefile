NAME=gitverse_notifier
BUILDDIR=build

VERSION ?= latest
BuildTime := $(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

LDFLAGS=-w -s

NOTIFIERLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
NOTIFIERLDFLAGS+=-X 'main.Githash=$(COMMIT)'
NOTIFIERLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'
NOTIFIERLDFLAGS+=-X 'main.Version=$(VERSION)'

NOTIFIERBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(NOTIFIERLDFLAGS) ${LDFLAGS}"

define make_artifact_full
	GOOS=$(1) GOARCH=$(2) $(NOTIFIERBUILD) -o $(BUILDDIR)/$(NAME)-$(1)-$(2)
	mkdir -p $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/locale/
	cp $(BUILDDIR)/$(NAME)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/$(NAME)
	cp README.md $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/README.md
	cp config.yml.example $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/config.yml.example
	cp entrypoint.sh $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)/entrypoint.sh

	cd $(BUILDDIR) && tar -czvf $(NAME)-$(VERSION)-$(1)-$(2).tar.gz $(NAME)-$(VERSION)-$(1)-$(2)
	rm -rf $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2) $(BUILDDIR)/$(NAME)-$(1)-$(2)
endef

build:
	echo "GOARCH=$(GOARCH) GOOS=$(GOOS) $(NOTIFIERBUILD) -o $(BUILDDIR)/$(NAME) ."
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(NOTIFIERBUILD) -o $(BUILDDIR)/$(NAME) .

all:
	$(call make_artifact_full,darwin,amd64)
	$(call make_artifact_full,darwin,arm64)
	$(call make_artifact_full,linux,amd64)
	$(call make_artifact_full,linux,arm64)
	$(call make_artifact_full,linux,mips64le)
	$(call make_artifact_full,linux,ppc64le)
	$(call make_artifact_full,linux,s390x)
	$(call make_artifact_full,linux,riscv64)
	$(call make_artifact_full,linux,loong64)

local:
	$(call make_artifact_full,$(shell go env GOOS),$(shell go env GOARCH))

darwin-amd64:
	$(call make_artifact_full,darwin,amd64)

darwin-arm64:
	$(call make_artifact_full,darwin,arm64)

linux-amd64:
	$(call make_artifact_full,linux,amd64)

linux-arm64:
	$(call make_artifact_full,linux,arm64)

linux-loong64:
	$(call make_artifact_full,linux,loong64)

linux-mips64le:
	$(call make_artifact_full,linux,mips64le)

linux-ppc64le:
	$(call make_artifact_full,linux,ppc64le)

linux-s390x:
	$(call make_artifact_full,linux,s390x)

linux-riscv64:
	$(call make_artifact_full,linux,riscv64)

.PHONY: docker
docker:
	@echo "build docker images"
	docker buildx build --build-arg VERSION=$(VERSION) -t gitverse-notifier:$(VERSION) . --load

.PHONY: clean
clean:
	-rm -rf $(BUILDDIR)
	-rm -rf $(UIDIR)/dist/*

.PHONY: run
run:
	go run ./gitverse_notifier
