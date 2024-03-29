WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries
BINARIES=dbgpProxy-macos dbgpProxy-macos-arm64 dbgpProxy-arm64 dbgpProxy-IBMi dbgpProxy dbgpProxy.exe

.NOTPARALLEL:

.PHONY: force

all: $(BINARIES) force

dbgpProxy: force
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

dbgpProxy-arm64: force
	GOOS=linux GOARCH=arm64 go build
	mv dbgpProxy dbgpProxy-arm64

dbgpProxy-macos: dbgpProxy-macos-arm64 force
	GOOS=darwin GOARCH=amd64 go build
	mv dbgpProxy dbgpProxy-macos

dbgpProxy-macos-arm64: force
	GOOS=darwin GOARCH=arm64 go build
	mv dbgpProxy dbgpProxy-macos-arm64

dbgpProxy-IBMi: force
	GOOS=aix GOARCH=ppc64 go build
	mv dbgpProxy dbgpProxy-IBMi

dbgpProxy.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done

deploy: all force
	$(eval VERSION := $(shell cat main.go | grep "var clientVersion" | sed 's/.*= "//' | sed 's/"//'))
	-git tag -s dbgpProxy-$(VERSION) -m "Go with dbgpProxy-$(VERSION)"
	cd $(WEBSITE_REPO) && git checkout master && git pull
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done
	cd $(BINARY_LOCATION) && git add $(BINARIES) && git commit -m "Added dbgpProxy binaries version $(VERSION)"
	cd $(WEBSITE_REPO) && git push
