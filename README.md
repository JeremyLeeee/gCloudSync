# gCloudSync
A cloud storage synchronizer implemented in Go.

## Quick start
### Config:
#### Example client config.json:
```json
{
    "ServerIP": "127.0.0.1",
    "TruncateBlockSize": 9192,
    "TransferBlockSize": 4096,
    "RootPath": "/Users/username/syncfolder"
}
```
#### Example server config.json:
```json
{
    "RootPath": "/Users/username/syncfolder"
}
```
where ServerIP represents server's public IP address. TruncateBlockSize represents the rsync block size used for checksum calculation. TransferBlockSize represents the max data package size sending through socket.

config.json should be placed in the same folder with executable binary.

### Build:
Go (version 1.17+) should be installed and added to path first.
#### Build binaries for all platform
```shell
./script/build.sh
```
#### Or run directly through go command
```shell
go run internal/client/main.go
```
or
```shell
go run internal/server/main.go
```
