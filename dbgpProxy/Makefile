.NOTPARALLEL:

.PHONY: force

all: dbgpProxy dbgpProxy-macos dbgpProxy.exe

dbgpProxy: dbgpProxy-macos force
	GOOS=linux GOARCH=amd64 go build

dbgpProxy-macos: force
	GOOS=darwin GOARCH=amd64 go build
	mv dbgpProxy dbgpProxy-macos

dbgpProxy.exe: force
	GOOS=windows GOARCH=amd64 go build
