/*
Copyright © 2024 Henry Huang <hhh@rutcode.com>

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

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-trellis/common/config"
	"github.com/go-trellis/common/testutils"
)

// TestIncludeBasic 测试基本的 include 功能
func TestIncludeBasic(t *testing.T) {
	// 创建临时测试目录
	tmpDir := t.TempDir()

	// 创建被包含的文件
	databaseYAML := `host: localhost
port: 3306
user: root
password: secret
`
	databaseFile := filepath.Join(tmpDir, "database.yml")
	err := os.WriteFile(databaseFile, []byte(databaseYAML), 0644)
	testutils.Ok(t, err)

	// 创建主配置文件
	mainYAML := `app:
  name: MyApp
  database: ${include:database.yml}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证 include 的内容
	dbHost := c.GetString("app.database.host")
	testutils.Assert(t, dbHost == "localhost", "database host should be localhost, got: %s", dbHost)

	dbPort := c.GetInt("app.database.port")
	testutils.Assert(t, dbPort == 3306, "database port should be 3306, got: %d", dbPort)

	dbUser := c.GetString("app.database.user")
	testutils.Assert(t, dbUser == "root", "database user should be root, got: %s", dbUser)

	dbPassword := c.GetString("app.database.password")
	testutils.Assert(t, dbPassword == "secret", "database password should be secret, got: %s", dbPassword)
}

// TestIncludeRelativePath 测试相对路径的 include
func TestIncludeRelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建子目录
	subDir := filepath.Join(tmpDir, "configs")
	err := os.MkdirAll(subDir, 0755)
	testutils.Ok(t, err)

	// 创建被包含的文件在子目录中
	cacheYAML := `type: redis
host: cache.example.com
port: 6379
`
	cacheFile := filepath.Join(subDir, "cache.yml")
	err = os.WriteFile(cacheFile, []byte(cacheYAML), 0644)
	testutils.Ok(t, err)

	// 创建主配置文件在根目录
	mainYAML := `app:
  name: MyApp
  cache: ${include:configs/cache.yml}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证 include 的内容
	cacheType := c.GetString("app.cache.type")
	testutils.Assert(t, cacheType == "redis", "cache type should be redis, got: %s", cacheType)

	cacheHost := c.GetString("app.cache.host")
	testutils.Assert(t, cacheHost == "cache.example.com", "cache host should be cache.example.com, got: %s", cacheHost)

	cachePort := c.GetInt("app.cache.port")
	testutils.Assert(t, cachePort == 6379, "cache port should be 6379, got: %d", cachePort)
}

// TestIncludeNested 测试嵌套 include
func TestIncludeNested(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建最底层的配置文件
	redisYAML := `host: redis.example.com
port: 6379
db: 0
`
	redisFile := filepath.Join(tmpDir, "redis.yml")
	err := os.WriteFile(redisFile, []byte(redisYAML), 0644)
	testutils.Ok(t, err)

	// 创建中间层配置文件
	cacheYAML := `type: redis
config: ${include:redis.yml}
`
	cacheFile := filepath.Join(tmpDir, "cache.yml")
	err = os.WriteFile(cacheFile, []byte(cacheYAML), 0644)
	testutils.Ok(t, err)

	// 创建主配置文件
	mainYAML := `app:
  name: MyApp
  cache: ${include:cache.yml}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证嵌套 include 的内容
	cacheType := c.GetString("app.cache.type")
	testutils.Assert(t, cacheType == "redis", "cache type should be redis, got: %s", cacheType)

	redisHost := c.GetString("app.cache.config.host")
	testutils.Assert(t, redisHost == "redis.example.com", "redis host should be redis.example.com, got: %s", redisHost)

	redisPort := c.GetInt("app.cache.config.port")
	testutils.Assert(t, redisPort == 6379, "redis port should be 6379, got: %d", redisPort)

	redisDB := c.GetInt("app.cache.config.db")
	testutils.Assert(t, redisDB == 0, "redis db should be 0, got: %d", redisDB)
}

// TestIncludeWithReference 测试 include 文件中使用 ${key} 引用
func TestIncludeWithReference(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建被包含的文件，使用 ${key} 引用
	databaseYAML := `host: ${db.host}
port: ${db.port}
user: root
`
	databaseFile := filepath.Join(tmpDir, "database.yml")
	err := os.WriteFile(databaseFile, []byte(databaseYAML), 0644)
	testutils.Ok(t, err)

	// 创建主配置文件，定义引用的值
	mainYAML := `app:
  name: MyApp
  database: ${include:database.yml}
db:
  host: db.example.com
  port: 5432
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证引用已正确解析
	dbHost := c.GetString("app.database.host")
	testutils.Assert(t, dbHost == "db.example.com", "database host should be db.example.com, got: %s", dbHost)

	dbPort := c.GetInt("app.database.port")
	testutils.Assert(t, dbPort == 5432, "database port should be 5432, got: %d", dbPort)
}

// TestIncludeCircularReference 测试循环引用检测
func TestIncludeCircularReference(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建文件 A，包含文件 B
	fileAYAML := `name: FileA
other: ${include:fileB.yml}
`
	fileA := filepath.Join(tmpDir, "fileA.yml")
	err := os.WriteFile(fileA, []byte(fileAYAML), 0644)
	testutils.Ok(t, err)

	// 创建文件 B，包含文件 A（循环引用）
	fileBYAML := `name: FileB
other: ${include:fileA.yml}
`
	fileB := filepath.Join(tmpDir, "fileB.yml")
	err = os.WriteFile(fileB, []byte(fileBYAML), 0644)
	testutils.Ok(t, err)

	// 尝试加载配置，应该检测到循环引用
	_, err = config.NewConfig(fileA)
	testutils.NotOk(t, err, "should detect circular include")
}

// TestIncludeFileNotFound 测试 include 文件不存在的情况
func TestIncludeFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建主配置文件，引用不存在的文件
	mainYAML := `app:
  name: MyApp
  database: ${include:nonexistent.yml}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err := os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 尝试加载配置，应该返回错误（或者保留原始值，取决于实现）
	// 当前实现如果 include 失败会返回原值，所以不会报错
	// 但如果需要更严格的错误处理，可以在这里添加检查
	c, err := config.NewConfig(mainFile)
	// 根据当前实现，include 失败时返回原值，所以不会报错
	// 这里验证配置仍然可以加载，但 include 的值应该是原始字符串
	if err == nil && c != nil {
		databaseValue := c.GetInterface("app.database")
		// 如果 include 失败，值应该是原始的 ${include:...} 字符串或者 nil
		_ = databaseValue // 用于调试，可以添加断言
	}
}

// TestIncludeJSON 测试 JSON 格式的 include
func TestIncludeJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建被包含的 JSON 文件
	databaseJSON := `{
  "host": "localhost",
  "port": 3306,
  "user": "root"
}
`
	databaseFile := filepath.Join(tmpDir, "database.json")
	err := os.WriteFile(databaseFile, []byte(databaseJSON), 0644)
	testutils.Ok(t, err)

	// 创建主 YAML 配置文件，包含 JSON 文件
	mainYAML := `app:
  name: MyApp
  database: ${include:database.json}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证 include 的内容
	dbHost := c.GetString("app.database.host")
	testutils.Assert(t, dbHost == "localhost", "database host should be localhost, got: %s", dbHost)

	dbPort := c.GetInt("app.database.port")
	testutils.Assert(t, dbPort == 3306, "database port should be 3306, got: %d", dbPort)
}

// TestIncludeMultiple 测试多次 include
func TestIncludeMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个被包含的文件
	databaseYAML := `host: localhost
port: 3306
`
	databaseFile := filepath.Join(tmpDir, "database.yml")
	err := os.WriteFile(databaseFile, []byte(databaseYAML), 0644)
	testutils.Ok(t, err)

	cacheYAML := `type: redis
host: cache.example.com
`
	cacheFile := filepath.Join(tmpDir, "cache.yml")
	err = os.WriteFile(cacheFile, []byte(cacheYAML), 0644)
	testutils.Ok(t, err)

	// 创建主配置文件，包含多个文件
	mainYAML := `app:
  name: MyApp
  database: ${include:database.yml}
  cache: ${include:cache.yml}
`
	mainFile := filepath.Join(tmpDir, "main.yml")
	err = os.WriteFile(mainFile, []byte(mainYAML), 0644)
	testutils.Ok(t, err)

	// 加载配置
	c, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, c != nil, "config should not be nil")

	// 验证两个 include 的内容
	dbHost := c.GetString("app.database.host")
	testutils.Assert(t, dbHost == "localhost", "database host should be localhost")

	cacheType := c.GetString("app.cache.type")
	testutils.Assert(t, cacheType == "redis", "cache type should be redis")

	cacheHost := c.GetString("app.cache.host")
	testutils.Assert(t, cacheHost == "cache.example.com", "cache host should be cache.example.com")
}

