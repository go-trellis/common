/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package fsm

import (
	"sync"

	"github.com/go-trellis/common/config"
)

type Config struct {
	// map[namespace]map[name]*Transition
	Namespaces map[string]*Namespace `json:"fsm" yaml:"fsm"`
}

type Namespace struct {
	Name        string        `json:"-" yaml:"-"`
	InitStatus  string        `json:"init_status" yaml:"init_status"`
	Transitions []*Transition `json:"transitions" yaml:"transitions"`
}

// NewFSMRepoFromConfigFile new Transitions from config file
func NewFSMRepoFromConfigFile(filepath string) (Repo, error) {
	cfg, err := config.NewConfigOptions(config.OptionFile(filepath))
	if err != nil {
		return nil, err
	}
	return NewRepo(cfg.GetConfig("fsm"))
}

// NewRepo new Transitions
func NewRepo(cfg config.Config) (Repo, error) {
	config := &Config{}
	if err := cfg.Object(config); err != nil {
		return nil, err
	}
	return newFSMFromConfig(config)
}

// newFSMFromConfig new FSM from config
func newFSMFromConfig(cfg *Config) (*FSM, error) {
	fsm := &FSM{transitions: make(map[string]*NamespaceTransitions), mu: &sync.RWMutex{}}
	for name, namespace := range cfg.Namespaces {
		if err := fsm.AddNamespace(name); err != nil {
			return nil, err
		}
		for _, transition := range namespace.Transitions {
			transition.Namespace = name
			if err := fsm.addTransition(transition); err != nil {
				return nil, err
			}
		}
		// set initial status for each namespace
		if err := fsm.SetCurrentStatus(name, namespace.InitStatus); err != nil {
			return nil, err
		}
	}

	return fsm, nil
}
