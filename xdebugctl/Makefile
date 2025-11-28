WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries
BINARIES=dbgpClient-arm64 dbgpClient dbgpClient.exe

.NOTPARALLEL:

.PHONY: force

all: xdebugctl xdebugctl.exe

xdebugctl: xdebugctl-arm64 force
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

xdebugctl-arm64: force
	GOOS=linux GOARCH=arm64 go build
	mv xdebugctl xdebugctl-arm64

xdebugctl.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done

deploy: all force
	$(eval VERSION := $(shell cat main.go | grep "var clientVersion" | sed 's/.*= "//' | sed 's/"//'))
	-git tag -s xdebugctl-$(VERSION) -m "Go with xdebugctl-$(VERSION)"
	cd $(WEBSITE_REPO) && git checkout master && git pull
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done
	cd $(BINARY_LOCATION) && git add $(BINARIES) && git commit -m "Added xdebugctl binaries version $(VERSION)"
	cd $(WEBSITE_REPO) && git push
