WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries

.NOTPARALLEL:

.PHONY: force

all: xdebug xdebug-macos xdebug.exe

xdebug: xdebug-macos xdebug-macos-arm64 xdebug-arm64 xdebug-IBMi force
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

xdebug-arm64: force
	GOOS=linux GOARCH=arm64 go build
	mv xdebug xdebug-arm64

xdebug-macos: xdebug-macos-arm64 force
	GOOS=darwin GOARCH=amd64 go build
	mv xdebug xdebug-macos

xdebug-macos-arm64: force
	GOOS=darwin GOARCH=arm64 go build
	mv xdebug xdebug-macos-arm64

xdebug-IBMi: force
	GOOS=aix GOARCH=ppc64 go build
	mv xdebug xdebug-IBMi

xdebug.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	cp xdebug $(BINARY_LOCATION)
	cp xdebug-arm64 $(BINARY_LOCATION)
	cp xdebug-macos $(BINARY_LOCATION)
	cp xdebug-macos-arm64 $(BINARY_LOCATION)
	cp xdebug-IBMi $(BINARY_LOCATION)
	cp xdebug.exe $(BINARY_LOCATION)
