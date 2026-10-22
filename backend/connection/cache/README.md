# Garbage collector command endpoint

Run the Go collector with the directory it is allowed to scan:

```powershell
go run ./backend/connection/cache/cmd/garbagecollector -dir "C:\path\to\data"
```

It listens on `127.0.0.1:8765` by default. The server accepts only `POST
/collect` with an empty body. It always scans the configured directory; clients
cannot provide a different path. The listener is restricted to loopback
addresses so the delete command is not exposed to other machines.

Trigger a run from Python:

```python
from backend.connection.cache.garbagecollector import collect

result = collect()
print(result["deleted"])
```

Trigger a run from Kotlin:

```kotlin
val resultJson = GarbageCollectorClient.collect()
println(resultJson)
```

The response is JSON with a `deleted` array. A non-success status or a
collection error is reported to the caller.
