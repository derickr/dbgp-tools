WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries
BINARIES=dbgpClient-macos dbgpClient-macos-arm64 dbgpClient-arm64 dbgpClient dbgpClient.exe

.NOTPARALLEL:

.PHONY: force

all: $(BINARIES) force

dbgpClient: force
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

dbgpClient-arm64: force
	GOOS=linux GOARCH=arm64 go build
	mv dbgpClient dbgpClient-arm64

dbgpClient-macos: dbgpClient-macos-arm64 force
	GOOS=darwin GOARCH=amd64 go build
	mv dbgpClient dbgpClient-macos

dbgpClient-macos-arm64: force
	GOOS=darwin GOARCH=arm64 go build
	mv dbgpClient dbgpClient-macos-arm64

dbgpClient.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done

deploy: all force
	$(eval VERSION := $(shell cat main.go | grep "var clientVersion" | sed 's/.*= "//' | sed 's/"//'))
	-git tag -s dbgpClient-$(VERSION) -m "Go with dbgpClient-$(VERSION)"
	cd $(WEBSITE_REPO) && git checkout master && git pull
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done
	cd $(BINARY_LOCATION) && git add $(BINARIES) && git commit -m "Added dbgpClient binaries version $(VERSION)"
	cd $(WEBSITE_REPO) && git push
