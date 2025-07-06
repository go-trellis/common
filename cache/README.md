# cache
light cache in go 

* [![GoDoc](http://godoc.org/trellis.tech/trellis/common.v2/cache?status.svg)](http://godoc.org/trellis.tech/trellis/common.v2/cache)

## Introduction

### Installation

```shell
go get "trellis.tech/trellis/common.v2/cache"
```

### Features

* Simple lru
* It can set Unique | Bag | DuplicateBag values per key

### TODO

* main node: to manage cache
* consistent hash to several nodes to install keys

#### Cache

cache is manager for k-vs tables base on TableCache

```go
// Cache Manager functions for executing k-v tables base on TableCache
type Cache interface {
	// Returns a list of all tables at the node.
	All() []string
	// Get TableCache
	GetTableCache(tab string) (TableCache, bool)
	// Creates a new table.
	New(tab string, options ...OptionFunc) error
	// Inserts the object or all of the objects in list.
	Insert(tab string, key, value any) bool
	// Inserts the object or all of the objects with expired time in list.
	InsertExpire(tab string, key, value any, expire time.Duration) bool
	// Deletes the entire table Tab.
	Delete(tab string) bool
	// Deletes all objects with key, Key from table Tab.
	DeleteObject(tab string, key any) bool
	// Delete all objects in the table Tab. Remain table in cache.
	DeleteObjects(tab string)
	// Look up values with key, Key from table Tab.
	Lookup(tab string, key any) ([]any, bool)
	// Look up all values in the Tab.
	LookupAll(tab string) (map[any][]any, bool)
	// Returns true if one or more elements in the table has key Key, otherwise false.
	Member(tab string, key any) bool
	// Retruns all keys in the table Tab.
	Members(tab string) ([]any, bool)
	// Set key Key expire time in the table Tab.
	SetExpire(tab string, key any, expire time.Duration) bool
}
```

#### TableCache

table cache is manager for k-vs

```golang
// TableCache
type TableCache interface {
	// Inserts the object or all of the objects in list.
	Insert(key, values any) bool
	// Inserts the object or all of the objects with expired time in list.
	InsertExpire(key, value any, expire time.Duration) bool
	// Deletes all objects with key: Key.
	DeleteObject(key any) bool
	// Delete all objects in the table Tab. Remain table in cache.
	DeleteObjects()
	// Returns true if one or more elements in the table has key: Key, otherwise false.
	Member(key any) bool
	// Retruns all keys in the table Tab.
	Members() ([]any, bool)
	// Look up values with key: Key.
	Lookup(key any) ([]any, bool)
	// Look up all values in the Tab.
	LookupAll() (map[any][]any, bool)
	// Set Key Expire time
	SetExpire(key any, expire time.Duration) bool
}
```

#### Sample: NewTableCache with options

[Examples](examples/main.go)
