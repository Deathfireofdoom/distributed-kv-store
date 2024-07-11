# Distributed KV-store

## Note 29-06-2024
The idea is to implement a distributed kv-store, using RAFT consensus algorithm, for now, its only a simple kv-store.


## Dev notes

* Fix so only leader accepts read/writes, maybe re-route otherwise?