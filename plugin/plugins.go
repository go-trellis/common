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
	"os"

	"trellis.tech/trellis/common.v0/config"
	"trellis.tech/trellis/common.v0/errcode"
)

var defPlugins Repo

func init() {
	plugins := NewPlugins()
	plugins.configFile = os.Getenv("TRELLIS_PLUGIN_FILENAME")
	defPlugins = plugins
	if err := defPlugins.Start(); err != nil {
		panic(err)
	}
}

func (p *Plugins) ParseFlags(set *flag.FlagSet) {
	set.StringVar(&p.configFile, "trellis.plugin.file-name", "", "plugin config path")
}

type Plugins struct {
	configFile    string
	trellisConfig config.Config

	configs Configs

	plugins map[string]Repo
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

func NewPlugins(opts ...Option) *Plugins {
	p := &Plugins{
		plugins: make(map[string]Repo),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Stop default plugins
func Stop() {
	defPlugins.Stop()
}

func (p *Plugins) Start() error {
	if p.configFile == "" && p.trellisConfig == nil {
		return nil
	}

	if p.configFile != "" {
		reader, err := config.NewSuffixReader(config.ReaderOptionFilename(p.configFile))
		if err != nil {
			return err
		}

		err = reader.Read(&p.configs)
		if err != nil {
			return err
		}
	}

	if p.trellisConfig != nil {
		if err := p.trellisConfig.ToObject("plugins", &p.configs); err != nil {
			return err
		}
	}

	for _, pConfig := range p.configs.Plugins {
		if _, ok := p.plugins[pConfig.Name]; ok {
			return errcode.Newf("config is already exists: %s", pConfig.Name)
		}
		plg, err := NewPlugin(pConfig)
		if err != nil {
			return err
		}
		p.plugins[pConfig.Name] = plg
	}

	for _, plg := range p.plugins {
		if err := plg.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugins) Stop() {
	for _, repo := range p.plugins {
		repo.Stop()
	}
}
