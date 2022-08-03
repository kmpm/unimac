

build:
	go build ./

install:
	copy unimac.exe c:\local\bin

unimac.exe: build

clean:
	-del unimac.exe
	-del clients.csv

clients: unimac.exe
	unimac.exe -o clients.csv clients


devices:
	go run ./ devices