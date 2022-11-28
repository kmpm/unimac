ifeq ($(OS),Windows_NT)
	FixPath = $(subst /,\,$1)
	BINEXT = .exe
else
	FixPath = $1
	BINEXT = 
endif

OUTPUT:=unimac$(BINEXT)
GIT_VERSION?=$(shell git describe --tags --always --long --dirty)
LDFLAGS:="-X 'main.appVersion=$(GIT_VERSION)'"
BUILDFLAGS:=-ldflags $(LDFLAGS)

.PHONY: build
build:
	go build $(BUILDFLAGS) -o $(OUTPUT) ./


$(OUTPUT): build


version:
	go run $(BUILDFLAGS) ./ version
	go run $(BUILDFLAGS) ./ licenses


clients:
	go run ./ clients -output clients.csv


devices:
	go run ./ devices -output devices.csv


.PHONY: clean
clean:
	-del $(OUTPUT)
	