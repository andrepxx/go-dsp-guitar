WIN_SYSROOT := $(HOME)/win-sysroot
CGO_FLAGS_AARCH64 := ""
CGO_FLAGS_ALLOW_WIN := ".*"
CGO_FLAGS_AMD64 := "-m64"
CGO_FLAGS_ARM := ""
CGO_FLAGS_I686 := "-m32"
CGO_FLAGS_WIN_AMD64 := "-m64 --sysroot=$(WIN_SYSROOT) -I$(WIN_SYSROOT)/include"
CGO_FLAGS_WIN_I686 := "-m32 --sysroot=$(WIN_SYSROOT) -I$(WIN_SYSROOT)/include"
CGO_LDFLAGS_ALLOW_WIN := ".*"
CGO_LDFLAGS_WIN := "--sysroot=$(WIN_SYSROOT) -L$(WIN_SYSROOT)/lib"
GCFLAGS_DEBUG := 'all=-N -l'
LDFLAGS_RELEASE := 'all=-w -s'
GOPATH := `pwd`/../../../..
PACKAGE_BIN := config ir keys webroot `ls dsp*`

all: dsp dsp-debug

.PHONY: clean clean-all fmt keys test

clean:
	rm -rf dist/
	rm -f dsp dsp-debug

clean-all:
	rm -rf dist/
	rm -f dsp dsp-debug dsp-linux-aarch64 dsp-linux-aarch64-debug dsp-linux-amd64 dsp-linux-amd64-debug dsp-linux-arm dsp-linux-arm-debug dsp-win-amd64.exe dsp-win-amd64-debug.exe dsp-win-i686.exe dsp-win-i686-debug.exe

dsp:
	GOPATH=$(GOPATH) go build -o dsp -ldflags $(LDFLAGS_RELEASE)

dsp-debug:
	GOPATH=$(GOPATH) go build -o dsp-debug -gcflags $(GCFLAGS_DEBUG)

dsp-linux-aarch64:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_AARCH64) CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 go build -o dsp-linux-aarch64 -ldflags $(LDFLAGS_RELEASE)

dsp-linux-aarch64-debug:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_AARCH64) CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 go build -o dsp-linux-aarch64-debug -gcflags $(GCFLAGS_DEBUG)

dsp-linux-amd64:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_AMD64) CC=x86_64-linux-gnu-gcc GOOS=linux GOARCH=amd64 go build -o dsp-linux-amd64 -ldflags $(LDFLAGS_RELEASE)

dsp-linux-amd64-debug:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_AMD64) CC=x86_64-linux-gnu-gcc GOOS=linux GOARCH=amd64 go build -o dsp-linux-amd64-debug -gcflags $(GCFLAGS_DEBUG)

dsp-linux-arm:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_ARM) CC=arm-linux-gnu-gcc GOOS=linux GOARCH=arm GOARM=7 go build -o dsp-linux-arm -ldflags $(LDFLAGS_RELEASE)

dsp-linux-arm-debug:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_ARM) CC=arm-linux-gnu-gcc GOOS=linux GOARCH=arm GOARM=7 go build -o dsp-linux-arm-debug -gcflags $(GCFLAGS_DEBUG)

dsp-win-amd64.exe:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_WIN_AMD64) CGO_LDFLAGS=$(CGO_LDFLAGS_WIN) CGO_CFLAGS_ALLOW=$(CGO_FLAGS_ALLOW_WIN) CGO_LDFLAGS_ALLOW=$(CGO_LDFLAGS_ALLOW_WIN) CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o dsp-win-amd64.exe -ldflags $(LDFLAGS_RELEASE)

dsp-win-amd64-debug.exe:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_WIN_AMD64) CGO_LDFLAGS=$(CGO_LDFLAGS_WIN) CGO_CFLAGS_ALLOW=$(CGO_FLAGS_ALLOW_WIN) CGO_LDFLAGS_ALLOW=$(CGO_LDFLAGS_ALLOW_WIN) CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o dsp-win-amd64-debug.exe -gcflags $(GCFLAGS_DEBUG)

dsp-win-i686.exe:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_WIN_I686) CGO_LDFLAGS=$(CGO_LDFLAGS_WIN) CGO_CFLAGS_ALLOW=$(CGO_FLAGS_ALLOW_WIN) CGO_LDFLAGS_ALLOW=$(CGO_LDFLAGS_ALLOW_WIN) CC=i686-w64-mingw32-gcc GOOS=windows GOARCH=386 go build -o dsp-win-i686.exe -ldflags $(LDFLAGS_RELEASE)

dsp-win-i686-debug.exe:
	GOPATH=$(GOPATH) CGO_ENABLED=1 CGO_CFLAGS=$(CGO_FLAGS_WIN_I686) CGO_LDFLAGS=$(CGO_LDFLAGS_WIN) CGO_CFLAGS_ALLOW=$(CGO_FLAGS_ALLOW_WIN) CGO_LDFLAGS_ALLOW=$(CGO_LDFLAGS_ALLOW_WIN) CC=i686-w64-mingw32-gcc GOOS=windows GOARCH=386 go build -o dsp-win-i686-debug.exe -gcflags $(GCFLAGS_DEBUG)

dist:
	mkdir dist
	mkdir dist/bin
	mkdir dist/bin/go-dsp-guitar
	cp -r $(PACKAGE_BIN) dist/bin/go-dsp-guitar/
	mkdir dist/src
	mkdir dist/src/go-dsp-guitar
	rsync -rlpv . dist/src/go-dsp-guitar/ --exclude dist/ --exclude ".*" --exclude "dsp*"
	cd dist/bin/ && tar cvzf go-dsp-guitar-vX.X.X.tar.gz --exclude=".[^/]*" go-dsp-guitar && cd ../../
	cd dist/src/ && tar cvzf go-dsp-guitar-vX.X.X.src.tar.gz --exclude=".[^/]*" go-dsp-guitar && cd ../../

fmt:
	GOPATH=$(GOPATH) gofmt -w .
	find \( -iname '*.css' -o -iname '*.js' -o -iname '*.json' -o -iname '*.md' -o -iname '*.xhtml' \) -execdir sed -i s/[[:space:]]*$$// {} \;

keys:
	mkdir keys
	openssl genrsa -out keys/private.pem 4096
	openssl req -new -x509 -days 365 -sha512 -key keys/private.pem -out keys/public.pem -subj "/C=DE/ST=Berlin/L=Berlin/O=None/OU=None/CN=localhost"

test:
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/circular
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/fft
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/level
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/oversampling
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/random
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/resample
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/tuner
	GOPATH=$(GOPATH) go test -cover github.com/andrepxx/go-dsp-guitar/wave

