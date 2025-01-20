# Build for Windows amd64
build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/sub2singbox.exe

# Build for Windows arm64 
build-windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o bin/sub2singbox.exe

# Build for Linux amd64
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/sub2singbox

# Build for Linux arm64
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/sub2singbox

# Build for macOS amd64
build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/sub2singbox

# Build for macOS arm64
build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/sub2singbox

# Build all platforms
build-all: build-windows-amd64 build-windows-arm64 build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64
