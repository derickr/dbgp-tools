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
