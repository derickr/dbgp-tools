WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries

.NOTPARALLEL:

.PHONY: force

all: dbgpClient dbgpClient-macos dbgpClient.exe

dbgpClient: dbgpClient-macos dbgpClient-macos-arm64 dbgpClient-arm64 force
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
	cp dbgpClient $(BINARY_LOCATION)
	cp dbgpClient-arm64 $(BINARY_LOCATION)
	cp dbgpClient-macos $(BINARY_LOCATION)
	cp dbgpClient-macos-arm64 $(BINARY_LOCATION)
	cp dbgpClient.exe $(BINARY_LOCATION)
