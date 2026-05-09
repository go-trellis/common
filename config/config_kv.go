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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-trellis/common.v2/errcode"
	"github.com/go-trellis/common.v2/files"
	"github.com/go-trellis/common.v2/types"
)

func (p *AdapterConfig) copyDollarSymbol(_ string, maps *map[string]any) error {
	for mapK, mapV := range *maps {
		value, err := p.checkValue(mapV)
		if err != nil {
			return err
		}
		if value != nil {
			(*maps)[mapK] = value
		}
	}
	return nil
}

func (p *AdapterConfig) checkValue(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32:
		return value, nil
	case reflect.Bool:
		return value, nil
	case reflect.String:
		s, ok := value.(string)
		if !ok {
			return value, nil
		}

		// 检查是否是 include 文件语法: ${include:filepath}
		includeRegex := regexp.MustCompile(includeFileReg)
		includeMatches := includeRegex.FindStringSubmatch(s)
		if len(includeMatches) > 1 {
			includePath := includeMatches[1] // 第一个捕获组是文件路径
			// 处理 include 文件
			includedConfig, err := p.processIncludeFile(includePath)
			if err != nil {
				// 对于循环引用等严重错误，返回错误
				return nil, err
			}
			return includedConfig, nil
		}

		// 检查是否是普通的 ${key} 语法
		_, matched := types.FindStringSubmatchMap(s, includeReg)
		if !matched {
			return value, nil
		}

		pKey := s[2 : len(s)-1]

		if p.EnvAllowed && (p.EnvPrefix == "" || strings.HasPrefix(pKey, p.EnvPrefix)) {
			if env := os.Getenv(pKey); env != "" {
				return env, nil
			}
		}
		v, err := p.getKeyValue(pKey)
		if err != nil {
			return nil, nil
		}

		return v, nil
	case reflect.Slice:
		vs := value.([]any)
		for i, v := range vs {
			newV, err := p.checkValue(v)
			if err != nil {
				return nil, err
			}
			vs[i] = newV
		}
		return vs, nil
	case reflect.Map:
		vs, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		for k, v := range vs {
			newV, err := p.checkValue(v)
			if err != nil {
				return nil, err
			}
			vs[k] = newV
		}
		return vs, nil
	default:
		panic(errcode.Newf("not suppered type: %+v", t))
	}
}

func (p *AdapterConfig) getKeyValue(key string) (any, error) {
	tokens := strings.Split(key, ".")
	vm := p.configs[tokens[0]]

	for i, t := range tokens {
		if i == 0 {
			continue
		}

		switch v := vm.(type) {
		case Options:
			vm = v[t]
		case map[string]any:
			vm = v[t]
		case map[any]any:
			vm = v[t]
		default:
			return nil, ErrNotMap
		}
	}
	return vm, nil
}

// processIncludeFile 处理 include 文件
func (p *AdapterConfig) processIncludeFile(includePath string) (any, error) {
	// 解析文件路径
	var absPath string
	if filepath.IsAbs(includePath) {
		// 即使是绝对路径，也规范化一下
		var err error
		absPath, err = filepath.Abs(includePath)
		if err != nil {
			return nil, errcode.Newf("cannot resolve include file path %q: %v", includePath, err)
		}
		absPath = filepath.Clean(absPath)
	} else {
		// 相对路径，基于当前配置文件的目录
		if len(p.ConfigFile) > 0 {
			// 确保 p.ConfigFile 是绝对路径
			baseConfigFile, err := filepath.Abs(p.ConfigFile)
			if err != nil {
				baseConfigFile = p.ConfigFile
			}
			baseDir := filepath.Dir(baseConfigFile)
			absPath = filepath.Join(baseDir, includePath)
			// 规范化路径
			absPath, err = filepath.Abs(absPath)
			if err != nil {
				return nil, errcode.Newf("cannot resolve include file path %q: %v", includePath, err)
			}
			absPath = filepath.Clean(absPath)
		} else {
			// 如果没有 ConfigFile，尝试使用工作目录
			var err error
			absPath, err = filepath.Abs(includePath)
			if err != nil {
				return nil, errcode.Newf("cannot resolve include file path %q: %v", includePath, err)
			}
			absPath = filepath.Clean(absPath)
		}
	}

	// 检查循环引用
	if p.includedFiles[absPath] {
		return nil, errcode.Newf("circular include detected for file %q", absPath)
	}

	// 标记文件为已包含
	p.includedFiles[absPath] = true
	defer func() {
		// 处理完成后，可以选择保留标记或移除
		// 这里保留标记，因为同一个文件可能在不同位置被 include
	}()

	// 读取 include 文件
	includeData, _, err := files.Read(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read include file %q: %w", absPath, err)
	}

	// 根据文件扩展名确定 reader type
	includeReaderType := fileToReaderType(absPath)
	var includeReader Reader

	switch includeReaderType {
	case ReaderTypeJSON:
		includeReader = NewJSONReader(ReaderOptionFilename(absPath))
	case ReaderTypeYAML:
		includeReader = NewYAMLReader(ReaderOptionFilename(absPath))
	default:
		return nil, fmt.Errorf("include file %q has unsupported format", absPath)
	}

	// 解析 include 文件内容
	var includeConfigs map[string]any
	if err := includeReader.ParseData(includeData, &includeConfigs); err != nil {
		return nil, fmt.Errorf("cannot parse include file %q: %w", absPath, err)
	}

	// 递归处理 include 文件中的 ${include:...} 和 ${key} 引用
	// 创建一个临时的 AdapterConfig 来处理 include 文件
	// 合并父配置的 configs，使得 include 文件可以引用主配置的值
	mergedConfigs := make(map[string]any)
	for k, v := range p.configs {
		mergedConfigs[k] = v
	}
	for k, v := range includeConfigs {
		mergedConfigs[k] = v
	}

	includeAdapter := &AdapterConfig{
		ConfigFile:    absPath,
		readerType:    includeReaderType,
		reader:        includeReader,
		configs:       mergedConfigs, // 使用合并后的 configs，可以访问父配置的值
		includedFiles: p.includedFiles, // 共享已包含文件列表
		EnvPrefix:     p.EnvPrefix,
		EnvAllowed:    p.EnvAllowed,
	}

	// 处理 include 文件中的 ${} 引用
	// 注意：只处理 includeConfigs，不要修改 mergedConfigs 中的父配置值
	if err := includeAdapter.copyDollarSymbol("", &includeConfigs); err != nil {
		return nil, fmt.Errorf("cannot process references in include file %q: %w", absPath, err)
	}

	return includeConfigs, nil
}

// setKeyValue set key value into *configs
func (p *AdapterConfig) setKeyValue(key string, value any) (err error) {
	tokens := strings.Split(key, ".")
	for i := len(tokens) - 1; i >= 0; i-- {
		if i == 0 {
			p.configs[tokens[0]] = value
			return
		}
		v, _ := p.getKeyValue(strings.Join(tokens[:i], "."))
		switch vm := v.(type) {
		case Options:
			vm[tokens[i]] = value
			value = vm
		case map[string]any:
			vm[tokens[i]] = value
			value = vm
		case map[any]any:
			vm[tokens[i]] = value
			value = vm
		default:
			value = map[string]any{tokens[i]: value}
		}
	}
	return
}
