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
	"regexp"
	"sort"
	"sync"
	"time"

	"trellis.tech/common.v2/files"

	"go.uber.org/zap/zapcore"
)

var (
	_ zapcore.WriteSyncer = (*fileLogger)(nil)
)

type fileLogger struct {
	options FileOptions

	mutex  sync.Mutex
	osFile *os.File

	backupFileReg *regexp.Regexp
}

// NewFileLogger 标准窗体的输出对象
func NewFileLogger(opts ...FileOption) (*fileLogger, error) {
	var options FileOptions
	for _, o := range opts {
		o(&options)
	}

	return NewFileLoggerWithOptions(options)
}

// NewFileLoggerWithOptions 标准窗体的输出对象
func NewFileLoggerWithOptions(opts FileOptions) (*fileLogger, error) {

	if err := opts.Check(); err != nil {
		return nil, err
	}

	fw := &fileLogger{
		options: opts,
	}

	err := fw.init()
	if err != nil {
		return nil, err
	}
	return fw, nil
}

func (p *fileLogger) init() (err error) {
	p.backupFileReg = regexp.MustCompile(fmt.Sprintf("%s_.*%s", p.options.FileBasename, p.options.FileExt))

	err = p.openFile()
	if err != nil {
		return
	}
	if p.options.Separator == "" {
		p.options.Separator = "\t"
	}

	if err = p.checkFile(0); err != nil {
		return err
	}

	return nil
}

func (p *fileLogger) Write(bs []byte) (int, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err := p.checkFile(int64(len(bs))); err != nil {
		return 0, err
	}
	return p.osFile.Write(bs)

}

func (p *fileLogger) Sync() error { return nil }

func (p *fileLogger) checkFile(dataLen int64) (err error) {

	if p.osFile == nil {
		err = p.openFile()
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		}
	}

	fi, err := p.osFile.Stat()
	if err != nil {
		return err
	}

	t := time.Now()

	if p.options.MoveFileType.getMoveFileFlag(t) == p.options.MoveFileType.getMoveFileFlag(fi.ModTime()) &&
		(p.options.MaxLength == 0 || (p.options.MaxLength > 0 && fi.Size()+dataLen < p.options.MaxLength)) {
		return nil
	}

	return p.moveFile(t)
}

func (p *fileLogger) openFile() (err error) {
	p.osFile, err = files.OpenWriteFile(filepath.Join(p.options.FileDir, p.options.Filename))
	return
}

func (p *fileLogger) moveFile(t time.Time) error {

	p.osFile = nil

	err := os.Rename(filepath.Join(p.options.FileDir, p.options.Filename),
		filepath.Join(p.options.FileDir, fmt.Sprintf("%s_%s%s",
			p.options.FileBasename, t.Format("20060102150405.999999"), p.options.FileExt)))
	if err != nil {
		return err
	}

	if err = p.removeOldFiles(); err != nil {
		return err
	}

	return p.openFile()
}

func (p *fileLogger) removeOldFiles() error {
	if p.options.MaxBackups == 0 {
		return nil
	}

	// 获取日志文件列表
	dirLis, err := ioutil.ReadDir(p.options.FileDir)
	if err != nil {
		return err
	}

	// 根据文件名过滤日志文件
	fileSort := FileSort{}
	//filePrefix := fmt.Sprintf("%s_", p.basename())
	for _, f := range dirLis {
		if p.backupFileReg.FindString(f.Name()) != "" {
			fileSort = append(fileSort, f)
		}
	}

	if fileSort.Len() <= p.options.MaxBackups {
		return nil
	}

	// 根据文件修改日期排序，保留最近的N个文件
	sort.Sort(fileSort)
	for _, f := range fileSort[p.options.MaxBackups:] {
		err := os.Remove(filepath.Join(p.options.FileDir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
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
