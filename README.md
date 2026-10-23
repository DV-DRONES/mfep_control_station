# mfep_control_station

## Garbage collector

Run the HTTP collector from the repository root, pointing `-dir` at the
specific directory whose JSON and YAML files should be scanned:

```powershell
go run .\backend\connection\wifi\cache\cmd\garbagecollector -dir C:\path\to\cache
```

Replace `C:\path\to\cache` with an existing directory. The service listens on
`127.0.0.1:8765`; send an empty `POST` request to
`http://127.0.0.1:8765/collect` to start a collection.

Do not run `go run` on `command.go`, `garbagecollector.go`, or a `*_test.go`
file. The first two belong to the `cache` library package, and Go test files
must be run with `go test`, for example:

```powershell
go test ./...
```

## Wi-Fi JSON/YAML transfer

The Wi-Fi transport sends `.json`, `.yaml`, or `.yml` files over TCP on the
local network. It stores each received file; it does not execute or interpret
its contents. The Go receiver handles one file per connection, up to 100 MiB,
and refuses to overwrite an existing file with the same name. Transfers are
not encrypted or authenticated, so use only on a trusted network.

Start the Go receiver from the repository root:

```powershell
go run .\backend\connection\wifi\cmd\wifi receive -listen 0.0.0.0:50000 -dir .\received
```

Use the receiver's Wi-Fi/LAN IP address in the sending program. Allow TCP port
`50000` through the receiver's firewall if needed. To accept connections only
from the same computer, use `-listen 127.0.0.1:50000`.

Send from Python:

```python
from backend.connection import wifi

wifi.send(r"C:\path\to\settings.yaml", host="192.168.1.20")
```

Send from Go:

```go
err := wifi.Send("192.168.1.20:50000", `C:\path\to\settings.yaml`)
```

The Go package is `mfep_control_station/backend/connection/wifi`; Go programs
can also send files with `wifi.Send("192.168.1.20:50000", filePath)`. Python
supports sending only; receiving is handled by the Go service.

```powershell
go run .\backend\connection\wifi\cmd\wifi send -address 192.168.1.20:50000 -file .\settings.yaml
```