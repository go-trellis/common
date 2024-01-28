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

package snowflake

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"trellis.tech/common.v2/crypto/base64"
	"trellis.tech/common.v2/errcode"
)

const (
	DefEpoch        int64 = 1504426600000 // 2017-09-03 16:16:40.000
	DefTimeAccuracy int64 = 1000000       // Millisecond 1000000
	DefMaxBits      uint8 = 63            // the max length of number (math:pow(2, 63)). without check generated id.
	DefNodeBits     uint8 = 10            // 1024 workers
	DefSequenceBits uint8 = 12            // 4096 numbers of per time accuracy
)

type ID int64

// String returns a string of the snowflake ID
func (p ID) String() string {
	return strconv.FormatInt(int64(p), 10)
}

// ParseString converts a string into a snowflake ID
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err
}

// Base2 returns a string base2 of the snowflake ID
func (p ID) Base2() string {
	return strconv.FormatInt(int64(p), 2)
}

// ParseBase2 converts a Base2 string into a snowflake ID
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// Base64 returns a base64 string of the snowflake ID
func (p ID) Base64() string {
	return base64.Encode(base64.EncodeStd, p.Bytes())
}

// ParseBase64 converts a base64 string into a snowflake ID
func ParseBase64(id string) (ID, error) {
	b, err := base64.DecodeString(base64.EncodeStd, id)
	if err != nil {
		return -1, err
	}
	return ParseBytes(b)
}

// Bytes returns a byte slice of the snowflake ID
func (p ID) Bytes() []byte {
	return []byte(p.String())
}

// ParseBytes converts a byte slice into a snowflake ID
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// Worker 工作对象
type Worker struct {
	locker sync.Mutex

	conf *Config

	lastTimestamp int64

	epoch time.Time

	sequence int64

	maxNodeID    int64
	sequenceMask int64

	nodeIDMask int64

	nodeIDShift uint8
	timeShift   uint8
}

type Option func(*Config)

// Config 配置
type Config struct {
	nodeID       int64
	epoch        int64
	epochDiff    int64
	maxBits      uint8
	sequenceBits uint8
	nodesBits    uint8
	timeAccuracy int64
}

func (p *Config) check() error {
	twEpochLen := len(strconv.Itoa(int(p.epoch)))
	switch twEpochLen {
	case 10: //秒
		//p.epoch *= 1000
		p.epochDiff = 1
		p.timeAccuracy = 1000000000
	case 13: //毫秒
		p.epochDiff = 1000
		p.timeAccuracy = 1000000
	case 16: //微秒
		//p.epoch /= 1000
		p.epochDiff = 1000000
		p.timeAccuracy = 1000
	default:
		return fmt.Errorf("eponch's length should be 10, 13 or 16")
	}

	if p.sequenceBits < 0 || p.nodesBits < 0 || p.maxBits < 0 || p.timeAccuracy < 0 {
		return errcode.New("bits can't less than 0")
	}

	if (p.sequenceBits + p.nodesBits) > p.maxBits {
		return errcode.Newf("sum of bits (%d) can't greater than the max bits(%d)",
			p.sequenceBits+p.nodesBits, p.maxBits)
	}
	return nil
}

func NodeID(id int64) Option {
	return func(c *Config) {
		c.nodeID = id
	}
}

func Epoch(epoch int64) Option {
	return func(c *Config) {
		c.epoch = epoch
	}
}

func MaxBits(maxBits uint8) Option {
	return func(c *Config) {
		c.maxBits = maxBits
	}
}

func SequenceBits(sequenceBits uint8) Option {
	return func(c *Config) {
		c.sequenceBits = sequenceBits
	}
}

func NodesBits(nodesBits uint8) Option {
	return func(c *Config) {
		c.nodesBits = nodesBits
	}
}

// NewWorker 生产工作对象
func NewWorker(opts ...Option) (*Worker, error) {

	c := &Config{
		epoch:        DefEpoch,
		timeAccuracy: DefTimeAccuracy,
		maxBits:      DefMaxBits,
		sequenceBits: DefSequenceBits,
		nodesBits:    DefNodeBits,
	}
	for _, o := range opts {
		o(c)
	}

	if err := c.check(); err != nil {
		return nil, err
	}

	w := &Worker{
		conf:          c,
		lastTimestamp: -1,
	}

	w.maxNodeID = -1 ^ (-1 << c.nodesBits)
	w.nodeIDMask = w.maxNodeID << c.sequenceBits
	w.sequenceMask = -1 ^ (-1 << c.sequenceBits)
	w.timeShift = c.sequenceBits + c.nodesBits
	w.nodeIDShift = c.sequenceBits

	if c.nodeID > w.maxNodeID || c.nodeID < 0 {
		return nil, fmt.Errorf("node id can't be greater than %d or less than 0", w.maxNodeID)
	}

	timeNow := time.Now()
	w.epoch = timeNow.Add(time.Unix(w.conf.epoch/w.conf.epochDiff,
		(w.conf.epoch%w.conf.epochDiff)*time.Second.Nanoseconds()/w.conf.timeAccuracy).
		Sub(timeNow))

	return w, nil
}

// Next 获取下一个ID值
// 统一时刻只能被调用一次
func (p *Worker) Next() ID {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.next(false)
}

// NextSleep 获取下一个ID值
// 时间戳一样，则沉睡1个单位时间
// 统一时刻只能被调用一次
func (p *Worker) NextSleep() ID {
	p.locker.Lock()
	defer p.locker.Unlock()
	return p.next(true)
}

func (p *Worker) next(sleep bool) ID {
	timestamp := p.timeGen()

	if p.lastTimestamp == timestamp {
		p.sequence = (p.sequence + 1) & p.sequenceMask
		if p.sequence == 0 {
			timestamp = p.timeGen()
			for timestamp <= p.lastTimestamp {
				if sleep {
					time.Sleep(time.Duration(p.conf.timeAccuracy))
				}
				timestamp = p.timeGen()
			}
		}
	} else {
		p.sequence = 0
	}

	p.lastTimestamp = timestamp

	return ID((timestamp << p.timeShift) | p.nodeIDMask | p.sequence)
}

// GetEpochTime 获取开始时间
func (p *Worker) GetEpochTime() int64 {
	return p.epoch.UnixNano()
}

func (p *Worker) tilNextTimestamp() int64 {
	timestamp := p.timeGen()
	for timestamp <= p.lastTimestamp {
		timestamp = p.timeGen()
	}
	return timestamp
}

func (p *Worker) timeGen() int64 {
	return time.Since(p.epoch).Nanoseconds() / p.conf.timeAccuracy
}
