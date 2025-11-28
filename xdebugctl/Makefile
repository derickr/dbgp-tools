WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries

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
	cp xdebugctl $(BINARY_LOCATION)
	cp xdebugctl-arm64 $(BINARY_LOCATION)
	cp xdebugctl.exe $(BINARY_LOCATION)
