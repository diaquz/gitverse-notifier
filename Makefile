NAME=gitverse_notifier
BUILDDIR=build

VERSION ?= latest
BuildTime := $(shell date -u '+%Y-%m-%d %I:%M:%S%p')
COMMIT ?= $(shell git rev-parse HEAD)
GOVERSION := $(shell go version)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

LDFLAGS=-w -s

NOTIFIERLDFLAGS+=-X 'main.Buildstamp=$(BuildTime)'
NOTIFIERLDFLAGS+=-X 'main.Githash=$(COMMIT)'
NOTIFIERLDFLAGS+=-X 'main.Goversion=$(GOVERSION)'
NOTIFIERLDFLAGS+=-X 'main.Version=$(VERSION)'

NOTIFIERBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(NOTIFIERLDFLAGS) ${LDFLAGS}"

init_configs:
	for file in *.yml.example; do \
		cp "$$file" "$$(basename "$$file" .example)"; \
	done

build:
	echo "GOARCH=$(GOARCH) GOOS=$(GOOS) $(NOTIFIERBUILD) -o $(BUILDDIR)/$(NAME) ."
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(NOTIFIERBUILD) -o $(BUILDDIR)/$(NAME) .

docker:
	docker buildx build --build-arg VERSION=$(VERSION) -t gitverse-notifier:$(VERSION) . --load

clean:
	-rm -rf $(BUILDDIR)

run:
	go run ./gitverse_notifier
