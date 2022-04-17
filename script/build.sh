# build amd64 client
echo "build client in amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -x -v -work -o ../build/amd64/linux/gCloudSync_client ../internal/client
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -x -v -work -o  ../build/amd64/darwin/gCloudSync_client ../internal/client
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -x -v -work -o  ../build/amd64/windows/gCloudSync_client.exe ../internal/client

# build arm64 client
echo "build client in arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -x -v -work -o  ../build/arm64/linux/gCloudSync_client ../internal/client
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -x -v -work -o  ../build/arm64/darwin/gCloudSync_client ../internal/client

# build amd64 server
echo "build server in amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -x -v -work -o  ../build/amd64/linux/gCloudSync_server ../internal/server
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -x -v -work -o  ../build/amd64/darwin/gCloudSync_server ../internal/server
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -x -v -work -o  ../build/amd64/windows/gCloudSync_server.exe ../internal/server

# build arm64 server
echo "build server in arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -x -v -work -o  ../build/arm64/linux/gCloudSync_server ../internal/server
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -x -v -work -o ../build/arm64/darwin/gCloudSync_server ../internal/server

echo "build finished."