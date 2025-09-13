/*
Copyright © 2023 Henry Huang <hhh@rutcode.com>

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

package logrushook

import (
	"fmt"
	"io"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"trellis.tech/trellis/common.v3/logger"
)

type LogrusConfig struct {
	Level         logrus.Level
	ReportCaller  bool
	DefaultWriter io.Writer
	Formatter     logrus.Formatter

	WriterInfo any
}

type TextFormatter = logrus.TextFormatter
type JSONFormatter = logrus.JSONFormatter

type LogrusLevelFileConfigs []*LogrusLevelFileConfig

type LogrusLevelFileConfig struct {
	Levels  []logrus.Level     `json:"levels" yaml:"levels"`
	Options logger.FileOptions `json:",inline" yaml:",inline"`
}

// NewLogrus 创建并配置一个新的logrus日志记录器
// 根据提供的LogrusConfig配置参数进行初始化
func NewLogrus(c *LogrusConfig) (*logrus.Logger, error) {
	// 参数校验
	if c == nil {
		return nil, fmt.Errorf("logrus config cannot be nil")
	}

	// 初始化基本日志记录器
	logrusLogger := initializeBaseLogger(c)

	// 如果没有配置WriterInfo，直接返回基本配置的日志记录器
	if c.WriterInfo == nil {
		return logrusLogger, nil
	}

	// 根据WriterInfo类型配置日志输出
	if err := configureLoggerOutput(logrusLogger, c.WriterInfo); err != nil {
		return nil, err
	}

	return logrusLogger, nil
}

// initializeBaseLogger 初始化基本的logrus日志记录器配置
func initializeBaseLogger(c *LogrusConfig) *logrus.Logger {
	logrusLogger := logrus.New()
	logrusLogger.SetReportCaller(c.ReportCaller)

	// 设置日志级别，如果配置的级别无效则使用警告级别
	if slices.Contains(logrus.AllLevels, c.Level) {
		logrusLogger.SetLevel(c.Level)
	} else {
		logrusLogger.SetLevel(logrus.WarnLevel)
	}

	// 设置格式化器
	if c.Formatter != nil {
		logrusLogger.SetFormatter(c.Formatter)
	}

	// 设置默认输出
	if c.DefaultWriter != nil {
		logrusLogger.SetOutput(c.DefaultWriter)
	}

	return logrusLogger
}

// configureLoggerOutput 根据WriterInfo类型配置日志输出
func configureLoggerOutput(logrusLogger *logrus.Logger, writerInfo interface{}) error {
	switch t := writerInfo.(type) {
	case string:
		// 简单的文件路径配置
		return handleStringWriter(logrusLogger, t)

	case *logger.FileOptions:
		// 单个文件选项配置
		return handleFileOptionsWriter(logrusLogger, *t)

	case map[logrus.Level]*logger.FileOptions:
		// 不同日志级别对应不同文件选项的映射
		return handleLevelFileOptionsMap(logrusLogger, t)

	case *LogrusLevelFileConfig:
		// 指针类型的日志级别文件配置
		return logrusWriterFileOptions(logrusLogger, t.Options, t.Levels)

	case LogrusLevelFileConfig:
		// 值类型的日志级别文件配置
		return logrusWriterFileOptions(logrusLogger, t.Options, t.Levels)

	case LogrusLevelFileConfigs:
		// 多个日志级别文件配置的集合
		return handleMultipleFileConfigs(logrusLogger, t)

	case []*LogrusLevelFileConfig:
		// 多个日志级别文件配置指针的切片
		return handleMultipleFileConfigs(logrusLogger, t)

	default:
		return fmt.Errorf("unsupported writer info type: %T", writerInfo)
	}
}

// handleStringWriter 处理字符串类型的文件路径配置
func handleStringWriter(logrusLogger *logrus.Logger, filePath string) error {
	flog, err := logger.NewFileLogger(logger.OptionFilename(filePath))
	if err != nil {
		return err
	}
	logrusWriterHook(logrusLogger, logrus.AllLevels, flog)
	return nil
}

// handleFileOptionsWriter 处理单个文件选项配置
func handleFileOptionsWriter(logrusLogger *logrus.Logger, options logger.FileOptions) error {
	flog, err := logger.NewFileLoggerWithOptions(options)
	if err != nil {
		return err
	}
	logrusWriterHook(logrusLogger, logrus.AllLevels, flog)
	return nil
}

// handleLevelFileOptionsMap 处理日志级别到文件选项的映射
func handleLevelFileOptionsMap(logrusLogger *logrus.Logger, levelOptionsMap map[logrus.Level]*logger.FileOptions) error {
	for level, options := range levelOptionsMap {
		if err := logrusWriterFileOptions(logrusLogger, *options, []logrus.Level{level}); err != nil {
			return err
		}
	}
	return nil
}

// handleMultipleFileConfigs 处理多个日志级别文件配置
func handleMultipleFileConfigs(logrusLogger *logrus.Logger, configs LogrusLevelFileConfigs) error {
	for _, config := range configs {
		if err := logrusWriterFileOptions(logrusLogger, config.Options, config.Levels); err != nil {
			return err
		}
	}
	return nil
}

func logrusWriterFileOptions(log *logrus.Logger, options logger.FileOptions, levels []logrus.Level) error {
	if len(levels) == 0 {
		return nil
	}
	flog, err := logger.NewFileLoggerWithOptions(options)
	if err != nil {
		return err
	}
	logrusWriterHook(log, levels, flog)
	return nil
}

func logrusWriterHook(logger *logrus.Logger, levels []logrus.Level, w io.Writer) {
	logger.AddHook(&writer.Hook{Writer: w, LogLevels: levels})
}
