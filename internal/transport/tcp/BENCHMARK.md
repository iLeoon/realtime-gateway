# The TCP server benchmark 


| Date | Component | Operation | Ops/sec | Allocs | B/op | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 2026-04-02 | **tcp/server** | **Connect** | **961k** | **10** | **564** | **Initial Baseline**|
| 2026-04-02 | tcp/server | Disconnect | 2809k | 5 | 143 | **Initial Baseline** |
| 2026-04-02 | tcp/server | Update | 10000000000000k | 0 | 0 | |




---

<details>
<summary><b>Detailed Log: 2026-04-02 | tcp/server</b></summary>

**Command:** `go test -bench=. -benchmem ./internal/transport/tcp`

**Raw Output:**
```text
</details>
goos: linux
goarch: amd64
pkg: github.com/iLeoon/realtime-gateway/internal/transport/tcp
cpu: Intel(R) Core(TM) i5-10500 CPU @ 3.10GHz
BenchmarkPacketsDispatcher/Connect-12            1000000              1087 ns/op             564 B/op         10 allocs/op
BenchmarkPacketsDispatcher/Disconnect-12         3394359               337.9 ns/op           143 B/op          5 allocs/op
BenchmarkPacketsDispatcher/Update-12            1000000000               0.0000001 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/iLeoon/realtime-gateway/internal/transport/tcp       2.611s

</details>
```

