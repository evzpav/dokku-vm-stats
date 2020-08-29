GO_REPO_ROOT := /go/src/github.com/evzpav/dokku-vm-stats
BUILD_IMAGE := golang:1.13

.PHONY: build-in-docker build clean src-clean

build-in-docker:
	docker run --rm \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "make build" || exit $$?

build: commands collectstats triggers
triggers: pre-deploy
commands: **/**/commands.go
	go build -a -o commands ./src/commands/*.go

collectstats: 
	go build -a -o collectstats ./src/scripts/collectstats.go

pre-deploy: **/**/pre-deploy.go
	go build -a -o pre-deploy ./src/triggers/pre-deploy.go

clean:
	rm -f commands pre-deploy scripts

src-clean:
	rm -rf .editorconfig .gitignore src LICENSE Makefile README.md *.go
