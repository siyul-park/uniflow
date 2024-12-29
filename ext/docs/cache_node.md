# Cache Node

**The Cache Node** implements a caching mechanism using an LRU (Least Recently Used) strategy to store and retrieve
data. This node provides caching capabilities to store results temporarily and reuses them for future requests,
improving performance by reducing the need for repeated processing.

## Specification

- **capacity**: Defines the maximum number of items the cache can hold. When the cache exceeds this capacity, the least
  recently used entries are evicted.
- **ttl**: Specifies the time-to-live (TTL) for cache entries. Once an entry expires, it will be removed from the cache.
  If not set, the cache does not have a TTL.

## Ports

- **in**: Receives the input packet and performs a cache lookup. If the data is found, it is returned. If not, the data
  is processed and added to the cache.
- **out**: Outputs the result of the cache lookup or processed data.

## Example

```yaml
- kind: cache
  capacity: 100   # Cache capacity of 100 items
  ttl: 1h         # Entries will expire after 1 hour
```
