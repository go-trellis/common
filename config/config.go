/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"math/big"
	"time"
)

// OptionFunc 处理函数
type OptionFunc func(*AdapterConfig)

// OptionFile 解析配置文件Option函数
func OptionFile(filename string) OptionFunc {
	return func(c *AdapterConfig) {
		c.ConfigFile = filename
	}
}

// OptionString 字符串解析配置Option函数
func OptionString(rt ReaderType, cStr string) OptionFunc {
	return func(c *AdapterConfig) {
		c.readerType = rt
		c.ConfigString = cStr
	}
}

// OptionStruct 结构体解析配置Option函数
func OptionStruct(rt ReaderType, st any) OptionFunc {
	return func(c *AdapterConfig) {
		c.readerType = rt
		c.ConfigStruct = st
	}
}

// OptionENVAllowed 允许获取系统环境变量
func OptionENVAllowed() OptionFunc {
	return func(c *AdapterConfig) {
		c.EnvAllowed = true
	}
}

// OptionENVPrefix 设置环境变量已自定义字符串开始
func OptionENVPrefix(prefix string) OptionFunc {
	return func(c *AdapterConfig) {
		c.EnvPrefix = prefix
	}
}

// Config manager data functions
type Config interface {
	// GetInterface get a object
	GetInterface(key string, defValue ...any) (res any)
	// GetString get a string
	GetString(key string, defValue ...string) (res string)
	// GetBoolean get a bool
	GetBoolean(key string, defValue ...bool) (b bool)
	// GetInt get a int
	GetInt(key string, defValue ...int) (res int)
	// GetFloat get a float
	GetFloat(key string, defValue ...float64) (res float64)
	// GetList get list of objects
	GetList(key string) (res []any)
	// GetStringList get list of strings
	GetStringList(key string) []string
	// GetBooleanList get list of bools
	GetBooleanList(key string) []bool
	// GetIntList get list of ints
	GetIntList(key string) []int
	// GetFloatList get list of float64s
	GetFloatList(key string) []float64
	// GetTimeDuration get time duration by (int)(uint), exp: 1s, 1day
	GetTimeDuration(key string, defValue ...time.Duration) time.Duration
	// GetByteSize get byte size by (int)(uint), exp: 1k, 1m
	GetByteSize(key string, defValue ...*big.Int) *big.Int
	// GetMap get map value
	GetMap(key string) Options
	// GetConfig get key's config
	GetConfig(key string) Config
	// ToObject unmarshal values to object
	// Deprecated: see function: Object
	ToObject(key string, model any) error
	// Object unmarshal values to object
	Object(model any, opts ...ObjOption) error
	// GetValuesConfig get key's values if values can be Config, or panic
	GetValuesConfig(key string) Config
	// SetKeyValue set key's value into config
	SetKeyValue(key string, value any) (err error)
	// Dump get all config
	Dump() (bs []byte, err error)
	// GetKeys get all keys
	GetKeys() []string
	// Copy deep copy configs
	Copy() Config
	IsEmpty() bool
}

type ObjOption func(*ObjOptions)

type ObjOptions struct {
	Key string
}

func ObjOptionKey(key string) ObjOption {
	return func(options *ObjOptions) {
		options.Key = key
	}
}

// NewConfig return Config by file's path, judge path's suffix, supported .json, .yml, .yaml
func NewConfig(name string) (Config, error) {
	if len(name) == 0 {
		return nil, ErrInvalidFilePath
	}
	return NewConfigOptions(OptionFile(name))
}

// NewConfigOptions 从操作函数解析Config
func NewConfigOptions(opts ...OptionFunc) (Config, error) {
	c := &AdapterConfig{}
	err := c.init(opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
