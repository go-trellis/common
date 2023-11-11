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
	"flag"
	"fmt"

	"github.com/go-kit/log"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/logger"
)

func (p *Plugins) ParseFlags(set *flag.FlagSet) {
	set.StringVar(&p.configFile, "trellis.plugin.file-name", "", "plugin config path")
}

type Plugins struct {
	configFile    string
	trellisConfig config.Config

	configs Configs
	workers map[string]Worker

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

// Logger set config logger
func Logger(l logger.KitLogger) Option {
	return func(p *Plugins) {
		p.logger = l
	}
}

func NewPlugins(opts ...Option) (*Plugins, error) {
	p := &Plugins{
		workers: make(map[string]Worker),
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

	for _, pConfig := range p.configs.Plugins {
		if err := p.registerWorker(pConfig); err != nil {
			return err
		}
	}

	for _, pConfig := range mapPluginConfigs {
		if err := p.registerWorker(pConfig); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugins) Start() error {
	for _, worker := range p.workers {
		if err := worker.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugins) registerWorker(c *Config) error {
	if _, ok := p.workers[c.Name]; ok {
		return errcode.Newf("plugin is already exists: %s", c.Name)
	}
	worker, err := NewPlugin(c, log.With(p.logger, "plugin", c.Name))
	if err != nil {
		return errcode.Newf("initial plugin failed: %+v", err)
	}
	p.workers[c.Name] = worker
	return nil
}

func (p *Plugins) Stop() {
	for _, worker := range p.workers {
		worker.Stop()
	}
}
