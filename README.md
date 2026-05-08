## Lightweight REST API with indexed Kaspa

---

### Environment

OS: 64-bit Linux (Ubuntu24.04 recommended).

HW: 4 Cores, 8GB RAM, 300GB SSD at least.

---

### Build the binary

Compile in the project directory.

<pre>go build</pre>

---

### Run using binary

A usable kaspa node for gRPC is required.

<pre>./kaspabook --kaspad-grpc=127.0.0.1:16110</pre>

---

### Show startup parameters:

<pre>
./kaspabook --help

KaspaBOOK v1.01.260430
Usage:
  kaspabook [OPTIONS]

Application Options:
      --showconfig      Show all configuration parameters.
      --hysteresis=     Number of DAA Scores hysteresis for data scanning. (default: 100)
      --concurrency=    Number of concurrent workers. (default: 8)
      --debug=          Debug information level; [0-3] available. (default: 2)
      --kaspad-grpc=    Kaspa node gRPC endpoint (comma-separated for multiple). (default: 127.0.0.1:16110)
      --rocks-path=     RocksDB data path. (default: ./data)
      --rocks-dtl=      Maximum DAA Score lifetime for indexed data. (default: 86400000)
      --rocks-gcloop    Enable proactive compaction loop. (default: true)
      --data-payload    Enable saving of transaction payload.
      --data-sigscript  Enable saving of transaction signature script.
      --api-host=       Listen host for the API server. (default: 0.0.0.0)
      --api-port=       Listen port for the API server. (default: 8003)
      --api-timeout=    Processing timeout for the API server in seconds. (default: 15)
      --api-connmax=    Maximum number of concurrent connections for the API server. (default: 1000)

Help Options:
  -h, --help          Show this help message
</pre>

---

### API Reference

https://github.com/5bb55b/go-kaspabook/blob/main/API.md

