# Config Reader

Go package for reading cofig file by JSON, XML, YAML.

## Installation

```bash
go get trellis.tech/trellis/common.v2/config
```

### imports

import [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml/tree/v3)

## Usage

### Config

> not supported "*.xml": now go encoding/xml is not supported map[string]any

* dot separator to get values, and if return nil, you should set default value
* A: ${X.Y.Z} for finding out X.Y.Z's value and setting into A. [See copy example](config_test.go#L20):[See config](example.json#14)
* You can do like this: c.GetString("a.b.c") Or c.GetString("a.b.c", "default")
* You can write notes into the json file.
* Supported: .json, .yaml

```go
c, e := NewConfig(name)
c.GetString("a.b.c")
```

### Feature

```go
// Config manager data functions
type Config interface {
	// get a object
	GetInterface(key string, defValue ...any) (res any)
	// get a string
	GetString(key string, defValue ...string) (res string)
	// get a bool
	GetBoolean(key string, defValue ...bool) (b bool)
	// get a int
	GetInt(key string, defValue ...int) (res int)
	// get a float
	GetFloat(key string, defValue ...float64) (res float64)
	// get list of objects
	GetList(key string) (res []any)
	// get list of strings
	GetStringList(key string) []string
	// get list of bools
	GetBooleanList(key string) []bool
	// get list of ints
	GetIntList(key string) []int
	// get list of float64s
	GetFloatList(key string) []float64
	// get time duration by (int)(uint), exp: 1s, 1day
	GetTimeDuration(key string, defValue ...time.Duration) time.Duration
	// get byte size by (int)(uint), exp: 1k, 1m
	GetByteSize(key string) *big.Int
	// get map value
	GetMap(key string) Options
	// get key's config
	GetConfig(key string) Config
	// get key's values if values can be Config, or panic
	GetValuesConfig(key string) Config
	// set key's value into config
	SetKeyValue(key string, value any) (err error)
	// get all config
	Dump() (bs []byte, err error)
	// get all keys
	GetKeys() []string
	// deep copy configs
	Copy() Config
}
```

### More Example

[See More Example]

* [JSON](example.json)
* [YAML](example.yml)

### Reader Repo

```go
// Reader reader repo
type Reader interface {
	// read file into model
	Read(model any) error
	// dump configs' cache
	Dump(model any) ([]byte, error)
	// parse data to model
	ParseData(data []byte, model any) error
}
```

```go
r := NewReader(ReaderType, filename)
if err := r.Read(model); err != nil {
	return
}
```

### Readers

```go
jReader := NewJSONReader() or NewJSONReader(ReaderOptionFilename(filename))
xReader := NewXMLReader()  or NewXMLReader(ReaderOptionFilename(filename))
yReader := NewYAMLReader() or NewYAMLReader(ReaderOptionFilename(filename))
```


> if not set filename with reader option function, you can't use Read function, it will return: no such file or directory

* .json = NewJSONReader() 
* .xml = NewXMLReader()
* .yaml | .yml = NewYAMLReader()

* if you want to use a fuzzy reader by filename's suffix

```go
sReader := NewSuffixReader(ReaderOptionFilename(filename))
```

