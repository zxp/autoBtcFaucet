tag_name = -X main.tagName=$(shell git describe --abbrev=0 --tags)
branch = -X main.branch=$(shell git rev-parse --abbrev-ref HEAD)
commit_id = -X main.commitID=$(shell git log --pretty=format:"%h" -1)
build_time = -X main.buildTime=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')

VERSION = $(tag_name) $(branch) $(commit_id) $(build_time)

.PHONY: all
.DEFAULT_GOAL := osx_amd64

osx_amd64:
	env GOOS=darwin GOARCH=amd64 go build -v -ldflags "-s -w ${VERSION}" -o builds/btcFaucet_osx
	cd getchats && env GOOS=darwin GOARCH=amd64 go build -v -ldflags "-s -w ${VERSION}" -o builds/getChats_osx
	mv getchats/builds/getChats_osx builds
	cd builds && 7z a btcFaucet_OSX.7z btcFaucet_osx getChats_osx

linux: linux_amd64 

linux_amd64:
	env GOOS=linux GOARCH=amd64 go build -v -ldflags "-s -w ${VERSION}" -o builds/btcFaucet_lin
	cd getchats && env GOOS=linux GOARCH=amd64 go build -v -ldflags "-s -w ${VERSION}" -o builds/getChats_lin
	mv getchats/builds/getChats_lin builds
	cd builds && 7z a btcFaucet_Linux.7z btcFaucet_lin getChats_lin

windows: windows_amd64

windows_amd64:
	env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CC_FOR_TARGET=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_CFLAGS="-I/Users/zxp/go/src/github.com/Arman92/go-tdlib/winlib -g -O2" CGO_LDFLAGS="-L/Users/zxp/go/src/github.com/Arman92/go-tdlib/winlib/td/build -ltdjson -g -O2" go build -x -v -ldflags "-s -w" -o builds/btcFaucet_x64.exe
	cd getchats && env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CC_FOR_TARGET=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 CGO_CFLAGS="-I/Users/zxp/go/src/github.com/Arman92/go-tdlib/winlib -g -O2" CGO_LDFLAGS="-L/Users/zxp/go/src/github.com/Arman92/go-tdlib/winlib/td/build -ltdjson -g -O2" go build -x -v -ldflags "-s -w" -o builds/getChats_x64.exe
	mv getchats/builds/getChats_x64.exe builds
	cd builds && 7z a btcFaucet_WinX64.7z *.dll *.exe

all: osx_amd64 linux windows
