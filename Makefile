
.PHONY: build
build:
	go build ./

install: unimac.exe
	copy unimac.exe c:\local\bin

unimac.exe: build

clean:
	-del unimac.exe
	-del clients.csv
	-del devices.csv


clients:
	go run ./ clients -output clients.csv


devices:
	go run ./ devices -output devices.csv