/*
Copyright © 2021 Henry Huang <hhh@rutcode.com>

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

package plugin

import (
	"os"
	"time"

	"trellis.tech/trellis/common.v0/errcode"
	"trellis.tech/trellis/common.v0/shell"
)

type Plugin struct {
	config Config

	ticker   *time.Ticker
	stopChan chan struct{}
}

func NewPlugin(config Config) (Repo, error) {
	if config.ScriptFile == "" {
		return nil, errcode.New("not set script file")
	}
	if _, err := os.Stat(config.ScriptFile); err != nil {
		return nil, errcode.NewErrors(
			errcode.Newf("not found script file: %s(%s)", config.Name, config.ScriptFile), err)
	}
	p := &Plugin{
		config: config,
	}
	if p.config.Interval > 0 {
		p.ticker = time.NewTicker(time.Duration(p.config.Interval))
		p.stopChan = make(chan struct{})
	}

	return p, nil
}

func (p *Plugin) Start() error {
	go p.do()
	return nil
}

func (p *Plugin) do() {
	intervalsGauge.WithLabelValues(p.config.Name, p.config.ScriptFile).
		Set(float64(time.Duration(p.config.Interval) / time.Second))
	if p.config.Interval <= 0 {
		startTime := time.Now()
		pluginLastDurationGauge.WithLabelValues(p.config.Name, p.config.ScriptFile).Set(float64(startTime.Unix()))
		evalTotalCounter.WithLabelValues(p.config.Name, p.config.ScriptFile).Add(1)
		err := shell.RunCommand(p.config.ScriptFile)
		if err != nil {
			evalFailureTotalCounter.WithLabelValues(p.config.Name, p.config.ScriptFile).Add(1)
		}
		lastTime := time.Now().Unix() - startTime.Unix()
		pluginExecuteSecondsGauge.WithLabelValues(p.config.Name, p.config.ScriptFile).Set(float64(lastTime))
		return
	}
	for {
		select {
		case t := <-p.ticker.C:
			pluginLastDurationGauge.WithLabelValues(p.config.Name, p.config.ScriptFile).Set(float64(t.Unix()))
			evalTotalCounter.WithLabelValues(p.config.Name, p.config.ScriptFile).Add(1)
			err := shell.RunCommand(p.config.ScriptFile)
			if err != nil {
				evalFailureTotalCounter.WithLabelValues(p.config.Name, p.config.ScriptFile).Add(1)
			}
			pluginExecuteSecondsGauge.WithLabelValues(p.config.Name, p.config.ScriptFile).
				Set(float64(time.Now().Unix() - t.Unix()))
		case <-p.stopChan:
			return
		}
	}
}

func (p *Plugin) Stop() {
	if p.stopChan != nil {
		p.stopChan <- struct{}{}
		close(p.stopChan)
	}
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
