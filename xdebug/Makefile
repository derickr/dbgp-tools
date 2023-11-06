.NOTPARALLEL:

.PHONY: force

all: xdebug xdebug-macos xdebug.exe

xdebug: xdebug-macos xdebug-macos-arm64 xdebug-arm64 force
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

xdebug.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	cp xdebug /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp xdebug-arm64 /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp xdebug-macos /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp xdebug-macos-arm64 /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp xdebug.exe /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
