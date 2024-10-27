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
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/robfig/cron/v3"

	"trellis.tech/trellis/common.v2/config"
	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/logger"
)

var mapPluginConfigs = map[string]*Config{}

func RegisterPlugin(name string, w func() error, opts ...OptionConfig) {
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

	if err := c.check(); err != nil {
		panic(err)
	}

	mapPluginConfigs[name] = c
}

func (p *Plugins) ParseFlags(set *flag.FlagSet) {
	set.StringVar(&p.configFile, "trellis.plugin.file-name", "", "plugin config path")
}

type Plugins struct {
	configFile    string
	trellisConfig config.Config
	cronOptions   []cron.Option

	configs   Configs
	plugins   map[string]*Plugin
	pluginIds map[cron.EntryID]*Plugin

	crons  *cron.Cron
	logger logger.KitLogger
}

type Configs struct {
	Plugins []*Config `yaml:"plugins" json:"plugins"`
}

type Option func(*Plugins)

// ConfigFile set config file
func ConfigFile(file string) Option {
	return func(p *Plugins) {
		p.configFile = file
	}
}

// TrellisConfig set config repo
func TrellisConfig(config config.Config) Option {
	return func(p *Plugins) {
		p.trellisConfig = config
	}
}

// CronOptions set cron options
func CronOptions(option ...cron.Option) Option {
	return func(p *Plugins) {
		p.cronOptions = append(p.cronOptions, option...)
	}
}

// Logger set config logger
func Logger(l logger.KitLogger) Option {
	return func(p *Plugins) {
		p.logger = l
	}
}

func NewPlugins(opts ...Option) (*Plugins, error) {
	p := &Plugins{
		plugins:   make(map[string]*Plugin),
		pluginIds: make(map[cron.EntryID]*Plugin),
	}
	for _, opt := range opts {
		opt(p)
	}
	if err := p.init(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Plugins) init() error {
	if p.logger == nil {
		p.logger = logger.Noop()
	}
	if p.configFile != "" {
		reader, err := config.NewSuffixReader(config.ReaderOptionFilename(p.configFile))
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = reader.Read(&p.configs)
		if err != nil {
			return err
		}
	}

	if p.trellisConfig != nil {
		if err := p.trellisConfig.Object(&p.configs, config.ObjOptionKey("plugins")); err != nil {
			return err
		}
	}

	p.crons = cron.New(p.cronOptions...)

	for _, pConfig := range p.configs.Plugins {
		if _, err := p.registerPlugin(pConfig); err != nil {
			return err
		}
	}

	for _, pConfig := range mapPluginConfigs {
		if _, err := p.registerPlugin(pConfig); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugins) registerPlugin(c *Config) (cron.EntryID, error) {
	if _, ok := p.plugins[c.Name]; ok {
		return 0, errcode.Newf("plugin is already exists: %s", c.Name)
	}
	plugin, err := NewPlugin(c, log.With(p.logger, "plugin", c.Name))
	if err != nil {
		return 0, errcode.Newf("initial plugin failed: %+v", err)
	}

	spec := ""
	if plugin.config.Interval > 0 {
		spec = fmt.Sprintf("@every %ds", plugin.config.Interval.Seconds())
		// spec = fmt.Sprintf("0/%d * * * *", plugin.config.Interval.Seconds())
	}

	if plugin.config.CronConfig != "" {
		spec = plugin.config.CronConfig
	}

	fmt.Println(c.Name, spec)

	id, err := p.crons.AddFunc(spec, p.runPlugin(plugin))
	if err != nil {
		return 0, err
	}

	p.plugins[c.Name] = plugin
	p.pluginIds[id] = plugin
	return id, nil
}

func (p *Plugins) runPlugin(plugin *Plugin) func() {
	return func() {
		t := time.Now()
		pluginLastDurationGauge.WithLabelValues(plugin.config.Name).Set(float64(t.Unix()))
		evalTotalCounter.WithLabelValues(plugin.config.Name).Add(1)

		if err := plugin.config.FN(); err != nil {
			level.Error(p.logger).Log("msg", "eval_function_failed", "error", err)
			evalFailureTotalCounter.WithLabelValues(plugin.config.Name).Add(1)
		} else {
			evalSuccessTotalCounter.WithLabelValues(plugin.config.Name).Add(1)
		}
		pluginExecuteSecondsSummary.WithLabelValues(plugin.config.Name).Observe(float64(time.Since(t) / 1e9))
	}
}

func (p *Plugins) Start() {
	p.crons.Start()
}

func (p *Plugins) Stop() context.Context {
	return p.crons.Stop()
}
