.NOTPARALLEL:

.PHONY: force

all: dbgpClient dbgpClient-macos dbgpClient.exe

dbgpClient: dbgpClient-macos force
	GOOS=linux GOARCH=amd64 go build

dbgpClient-macos: force
	GOOS=darwin GOARCH=amd64 go build
	mv dbgpClient dbgpClient-macos

dbgpClient.exe: force
	GOOS=windows GOARCH=amd64 go build

update-website: all
	cp dbgpClient /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient-macos /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
	cp dbgpClient.exe /home/derick/dev/php/xdebug-xdebug.org/html/files/binaries
