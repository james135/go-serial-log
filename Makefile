build_mac:
	GOOS=darwin GOARCH=amd64 go build  -o ./builds/refurbination

build_linux:
	GOOS=linux GOARCH=amd64 go build -o ./builds/refurbination

build_silicon:
	GOOS=darwin GOARCH=arm64 go build -o ./builds/refurbination
	
build_pi5:
	GOOS=linux GOARCH=arm64 go build -o ./builds/refurbination