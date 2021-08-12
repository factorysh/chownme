build: client server

client: bin
	GOOS=linux go build -o bin/chownme cli/chownme/main.go

server: bin
	GOOS=linux go build -o bin/chownmed cli/chownmed/main.go

bin:
	mkdir -p bin
