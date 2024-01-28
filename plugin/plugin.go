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
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"trellis.tech/common.v2/errcode"
	"trellis.tech/common.v2/logger"
	"trellis.tech/common.v2/shell"
	"trellis.tech/common.v2/types"
)

var (
	evalFailureTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "plugin_evaluation_failures_total",
			Help:      "The total number of plugin evaluation failures.",
		},
		[]string{"plugin"},
	)
	evalSuccessTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "plugin_evaluation_success_total",
			Help:      "The total number of plugin evaluation success.",
		},
		[]string{"plugin"},
	)
	evalTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "plugin_evaluation_total",
			Help:      "The evaluation total number of the workers.",
		},
		[]string{"plugin"},
	)
	intervalsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trellis",
			Name:      "plugin_interval_seconds",
			Help:      "The interval of a plugin.",
		},
		[]string{"plugin"},
	)
	pluginExecuteSecondsSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "trellis",
			Name:       "plugin_last_duration_seconds",
			Help:       "The time of the last plugin evaluation.",
			Objectives: map[float64]float64{0.1: 0.09, 0.5: 0.05, 0.75: 0.025, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"plugin"},
	)
	pluginLastDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trellis",
			Name:      "plugin_last_evaluation_timestamp_seconds",
			Help:      "The timestamp of the last plugin evaluation in seconds.",
		},
		[]string{"plugin"},
	)
)

func init() {
	prometheus.MustRegister(
		evalFailureTotalCounter,
		evalSuccessTotalCounter,
		evalTotalCounter,
		intervalsGauge,
		pluginExecuteSecondsSummary,
		pluginLastDurationGauge,
	)
}

type Worker interface {
	Start() error
	Stop()
}

type Plugin struct {
	config *Config
	logger logger.KitLogger

	ticker   *time.Ticker
	stopChan chan struct{}
}

type Config struct {
	Name       string         `yaml:"name" json:"name"`
	ScriptFile string         `yaml:"script_file" json:"script_file"`
	ScriptArgs []string       `yaml:"script_args" json:"script_args"`
	Interval   types.Duration `yaml:"interval" json:"interval"`

	FN func() error `yaml:"-" json:"-"`
}

func (p *Config) check() error {
	if p == nil {
		return fmt.Errorf("nil config")
	}
	if p.Name == "" {
		return fmt.Errorf("empty plugin's name")
	}
	if p.Interval < 0 {
		p.Interval = 0
	}
	return nil
}

type ConfigOption = func(*Config)

func Interval(i time.Duration) ConfigOption {
	interval := types.Duration(i)
	if interval <= 0 {
		interval = types.Duration(time.Minute)
	}
	return func(c *Config) {
		c.Interval = interval
	}
}

var mapPluginConfigs = map[string]*Config{}

func RegisterPlugin(name string, w func() error, opts ...ConfigOption) {
	_, ok := mapPluginConfigs[name]
	if ok {
		panic(fmt.Errorf("plugin already exist: %s", name))
	}
	c := &Config{
		Name: name,
		FN:   w,
	}
	for _, opt := range opts {
		opt(c)
	}

	mapPluginConfigs[name] = c
}

func NewPlugin(c *Config, l logger.KitLogger) (Worker, error) {
	if c.ScriptFile == "" && c.FN == nil {
		return nil, errcode.New("not set script file or function")
	}
	if c.ScriptFile != "" {
		if _, err := os.Stat(c.ScriptFile); err != nil {
			return nil, errcode.NewErrors(
				errcode.Newf("not found script file: %s(%s)", c.Name, c.ScriptFile), err)
		}
		c.FN = func() error {
			return shell.RunCommand(c.ScriptFile, c.ScriptArgs...)
		}
	}

	if l == nil {
		l = logger.Noop()
	}

	p := &Plugin{
		config: c,
		logger: l,
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
	intervalsGauge.WithLabelValues(p.config.Name).Set(float64(time.Duration(p.config.Interval) / time.Second))
	p.doRun(time.Now())
	if p.config.Interval <= 0 {
		return
	}
	for {
		select {
		case t := <-p.ticker.C:
			p.doRun(t)
		case <-p.stopChan:
			return
		}
	}
}

func (p *Plugin) doRun(t time.Time) {
	pluginLastDurationGauge.WithLabelValues(p.config.Name).Set(float64(t.Unix()))
	evalTotalCounter.WithLabelValues(p.config.Name).Add(1)
	err := p.config.FN()
	if err != nil {
		level.Error(p.logger).Log("msg", "eval_function_failed", "error", err)
		evalFailureTotalCounter.WithLabelValues(p.config.Name).Add(1)
	} else {
		evalSuccessTotalCounter.WithLabelValues(p.config.Name).Add(1)
	}
	pluginExecuteSecondsSummary.WithLabelValues(p.config.Name).Observe(float64(time.Since(t) / 1e9))
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
