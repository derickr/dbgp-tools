.NOTPARALLEL:

.PHONY: force

all: dbgpClient dbgpClient-macos dbgpClient.exe

dbgpClient: dbgpClient-macos dbgpClient-macos-arm64 dbgpClient-arm64 force
	GOOS=linux GOARCH=amd64 go build

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
	cp dbgpClient /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient-arm64 /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient-macos /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient-macos-arm64 /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient.exe /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
