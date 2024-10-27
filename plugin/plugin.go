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

	"github.com/prometheus/client_golang/prometheus"

	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/logger"
	"trellis.tech/trellis/common.v2/shell"
	"trellis.tech/trellis/common.v2/types"
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

func OptionCronConfig(cronConfig string) OptionConfig {
	return func(c *Config) {
		c.CronConfig = cronConfig
	}
}

func OptionInterval(interval types.Duration) OptionConfig {
	return func(c *Config) {
		c.Interval = interval
	}
}

type OptionConfig func(*Config)

type Plugin struct {
	config *Config
	logger logger.KitLogger
}

type Config struct {
	Name       string   `yaml:"name" json:"name"`
	ScriptFile string   `yaml:"script_file" json:"script_file"`
	ScriptArgs []string `yaml:"script_args" json:"script_args"`

	CronConfig string         `yaml:"cron_config" json:"cron_config"`
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

	if p.ScriptFile == "" && p.FN == nil {
		return errcode.New("not set script file or function")
	}

	if p.Interval <= 0 && p.CronConfig == "" {
		return fmt.Errorf("not set cron config or executor interval")
	}
	return nil
}

func NewPlugin(c *Config, l logger.KitLogger) (*Plugin, error) {
	if err := c.check(); err != nil {
		return nil, err
	}
	p := &Plugin{
		config: c,
		logger: l,
	}

	if p.logger == nil {
		p.logger = logger.Noop()
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

	return p, nil
}
