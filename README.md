# Distributed KV-store

## Note 29-06-2024
The idea is to implement a distributed kv-store, using RAFT consensus algorithm, for now, its only a simple kv-store.


## Dev notes

* Fix so the service does not panic when the grpc is not answering
* Fix so a node can come back after failure