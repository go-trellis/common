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
	"github.com/prometheus/client_golang/prometheus"
	"trellis.tech/trellis/common.v0/types"
)

type Repo interface {
	Start() error
	Stop()
}

type Configs struct {
	Plugins []Config `yaml:"plugins" json:"plugins"`
}

type Config struct {
	Name       string         `yaml:"name" json:"name"`
	ScriptFile string         `yaml:"script_file" json:"script_file"`
	Interval   types.Duration `yaml:"interval" json:"interval"`
}

var (
	evalFailureTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "common_plugin_evaluation_failures_total",
			Help:      "The total number of plugin evaluation failures.",
		},
		[]string{"plugin", "file_path"},
	)
	evalTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trellis",
			Name:      "common_plugin_evaluation_total",
			Help:      "The evaluation total number of the plugins.",
		},
		[]string{"plugin", "file_path"},
	)
	intervalsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trellis",
			Name:      "common_plugin_interval_seconds",
			Help:      "The interval of a plugin.",
		},
		[]string{"plugin", "file_path"},
	)
	pluginExecuteSecondsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trellis",
			Name:      "common_plugin_last_duration_seconds",
			Help:      "The time of the last plugin evaluation.",
		},
		[]string{"plugin", "file_path"},
	)
	pluginLastDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "trellis",
			Name:      "common_plugin_last_evaluation_timestamp_seconds",
			Help:      "The timestamp of the last plugin evaluation in seconds.",
		},
		[]string{"plugin", "file_path"},
	)
)

func init() {
	prometheus.MustRegister(evalFailureTotalCounter)
	prometheus.MustRegister(evalTotalCounter)
	prometheus.MustRegister(intervalsGauge)
	prometheus.MustRegister(pluginExecuteSecondsGauge)
	prometheus.MustRegister(pluginLastDurationGauge)
}
