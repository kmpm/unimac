ifeq ($(OS),Windows_NT)
	FixPath = $(subst /,\,$1)
	BINEXT = .exe
	cp = copy $(1) $(2)
else
	FixPath = $1
	BINEXT = 
	cp = cp $1 $2
endif

OUTPUT:=unimac$(BINEXT)
GIT_VERSION?=$(shell git describe --tags --always --long --dirty)
LDFLAGS:="-X 'main.appVersion=$(GIT_VERSION)'"
BUILDFLAGS:=-ldflags $(LDFLAGS)

.PHONY: build
build: out
	go build $(BUILDFLAGS) -o $(OUTPUT) ./
	$(call cp,$(OUTPUT),out/$(GIT_VERSION)-$(OUTPUT))

out:
	mkdir out

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
	
today: out
	go run ./ clients -output out/today/clients.csv
	go run ./ devices -output out/today/devices.csv