.PHONY: all clean build-x64 build-x86 build-arm64 build-armv7

APP_NAME=go-ipxe-admin
BUILD_DIR=bin

all: clean build-x64 build-x86 build-arm64 build-armv7

clean:
	rm -rf $(BUILD_DIR)

build-x64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go
	upx -9 $(BUILD_DIR)/$(APP_NAME)-linux-amd64

build-x86:
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-386 main.go
	upx -9 $(BUILD_DIR)/$(APP_NAME)-linux-386

build-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 main.go
	upx -9 $(BUILD_DIR)/$(APP_NAME)-linux-arm64

build-armv7:
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 main.go
	upx -9 $(BUILD_DIR)/$(APP_NAME)-linux-armv7

dev:
	ADMIN_USER=admin ADMIN_PASS=admin go run main.go
