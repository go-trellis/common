/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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

package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
	"trellis.tech/trellis/common.v0/errcode"
	"trellis.tech/trellis/common.v0/files"
)

var (
	_ zapcore.WriteSyncer = (*fileLogger)(nil)
)

type fileLogger struct {
	options FileOptions

	mutex    sync.Mutex
	fileRepo files.FileRepo
}

// NewFileLogger 标准窗体的输出对象
func NewFileLogger(opts ...FileOption) (*fileLogger, error) {
	var options FileOptions
	for _, o := range opts {
		o(&options)
	}

	fw, err := NewFileLoggerWithOptions(options)
	if err != nil {
		return nil, err
	}

	return fw, err
}

// NewFileLoggerWithOptions 标准窗体的输出对象
func NewFileLoggerWithOptions(opts FileOptions) (*fileLogger, error) {
	fw := &fileLogger{
		options: opts,
	}

	err := fw.init()
	if err != nil {
		return nil, err
	}
	return fw, nil
}

func (p *fileLogger) init() error {
	if p == nil || p.options.Filename == "" {
		return errcode.New("file name not exist")
	}

	p.fileRepo = files.NewFileRepo(files.ConcurrencyRead())
	if p.options.Separator == "" {
		p.options.Separator = "\t"
	}

	if err := p.checkFile(); err != nil {
		return err
	}

	return nil
}

func (p *fileLogger) Write(bs []byte) (int, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err := p.checkFile(); err != nil {
		return 0, err
	}
	return p.fileRepo.WriteAppendBytes(p.options.Filename, bs)
}

func (p *fileLogger) Sync() error { return nil }

func (p *fileLogger) checkFile() (err error) {

	fi, err := p.fileRepo.FileInfo(p.options.Filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	t := time.Now()

	if p.options.MoveFileType.getMoveFileFlag(t) == p.options.MoveFileType.getMoveFileFlag(fi.ModTime()) &&
		(p.options.MaxLength == 0 || (p.options.MaxLength > 0 && fi.Size() < p.options.MaxLength)) {
		return nil
	}

	return p.moveFile(t)
}

func (p *fileLogger) moveFile(t time.Time) error {

	err := p.fileRepo.Rename(
		p.options.Filename, fmt.Sprintf("%s_%s", p.options.Filename, t.Format("20060102150405.999999999")))
	if err != nil {
		return err
	}

	if err = p.removeOldFiles(); err != nil {
		return err
	}

	_, err = p.fileRepo.Write(p.options.Filename, "")

	return err
}

func (p *fileLogger) removeOldFiles() error {
	if p.options.MaxBackups == 0 {
		return nil
	}

	// 获取日志文件列表
	dirLis, err := ioutil.ReadDir(p.dir())
	if err != nil {
		return err
	}

	// 根据文件名过滤日志文件
	fileSort := FileSort{}
	filePrefix := fmt.Sprintf("%s_", p.basename())
	for _, f := range dirLis {
		if strings.Contains(f.Name(), filePrefix) {
			fileSort = append(fileSort, f)
		}
	}

	if fileSort.Len() <= p.options.MaxBackups {
		return nil
	}

	// 根据文件修改日期排序，保留最近的N个文件
	sort.Sort(fileSort)
	for _, f := range fileSort[p.options.MaxBackups:] {
		err := os.Remove(filepath.Join(p.dir(), f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// dir returns the directory for the current filename.
func (p *fileLogger) dir() string {
	return filepath.Dir(p.options.Filename)
}

// filename generates the name of the logfile from the current time.
func (p *fileLogger) basename() string {
	return filepath.Base(p.options.Filename)
}

// FileSort 文件排序
type FileSort []os.FileInfo

func (fs FileSort) Len() int {
	return len(fs)
}

func (fs FileSort) Less(i, j int) bool {
	return fs[i].Name() > fs[j].Name()
}

func (fs FileSort) Swap(i, j int) {
	fs[i], fs[j] = fs[j], fs[i]
}
