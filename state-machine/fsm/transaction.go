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

	"github.com/go-trellis/common.v3/errors/errcode"
)

var (
	_ TransitionRepo = (*NamespaceTransitions)(nil)
)

// Transition information for current to target status in namespace
type Transition struct {
	// namespace of transition
	Namespace string `json:"-" yaml:"-"`
	// current status in namespace that the transition will start from triggered by event
	CurrentStatus string `json:"current" yaml:"current"`
	// event that trigger the transition from current to target status in namespace
	Event string `json:"event" yaml:"event"`
	// target status in namespace that the transition will reach from current status triggered by event
	TargetStatus string `json:"target" yaml:"target"`
	// description of the transition, optional field for documentation purposes
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type NamespaceTransitions struct {
	// namespace of transition
	Namespace string `json:"namespace" yaml:"namespace"`
	// current status in namespace that the transition will start from triggered by event
	CurrentStatus string `json:"current" yaml:"current"`
	// map of current to target status in namespace that the transition will start from triggered by event
	Transitions map[string]*Transition `json:"transitions" yaml:"transitions"`

	rwMux sync.RWMutex `json:"-" yaml:"-"`
}

// valid
func (p *Transition) valid() error {
	if p == nil {
		return errcode.New("nil transition")
	}

	if p.Namespace == "" {
		return errcode.New("missing namespace")
	}

	return p.validWithNamespace(p.Namespace)
}

// validWithNamespace checks if the transition is valid with the given namespace. It updates the namespace if necessary and checks for empty strings in the transition fields. If any of the fields are invalid, it returns an error.
func (p *Transition) validWithNamespace(namespace string) error {
	if p == nil {
		return errcode.New("nil transition")
	}

	if p.Namespace == "" {
		p.Namespace = namespace
	} else if p.Namespace != namespace {
		return errcode.Newf("namespace mismatch, expected %s but got %s", p.Namespace, namespace)
	}

	// check if namespace, event and current status are empty strings
	if p.Event == "" || p.CurrentStatus == "" {
		return errcode.New("invalid transition")
	}

	// check if target status is empty string
	if p.TargetStatus == "" {
		return errcode.New("empty target status")
	}

	return nil
}

// AddTransaction adds a transition to the namespace transitions. If the transition is already present, it returns an error.
func (p *NamespaceTransitions) AddTransaction(trans *Transition) error {
	if err := trans.validWithNamespace(p.Namespace); err != nil {
		return err
	}

	key := genKey(trans.CurrentStatus, trans.Event)
	p.rwMux.Lock()
	defer p.rwMux.Unlock()
	if _, ok := p.Transitions[key]; ok {
		return errcode.Newf("duplicate transition: %s, %s", trans.CurrentStatus, trans.Event)
	}
	p.Transitions[key] = trans
	return nil
}

// RemoveTransition removes a transition from the namespace transitions. If the transition is not present, it returns an error.
func (p *NamespaceTransitions) RemoveTransition(status, event string) error {
	key := genKey(status, event)
	p.rwMux.Lock()
	defer p.rwMux.Unlock()
	if _, ok := p.Transitions[key]; !ok {
		return errcode.Newf("no such transition: %s, %s", status, event)
	}
	delete(p.Transitions, key)
	return nil
}

// ChangeStatus changes the current status based on the event. If the event is not found in the transitions, it returns an error.
func (p *NamespaceTransitions) ChangeStatus(event string) (string, error) {
	p.rwMux.Lock()
	defer p.rwMux.Unlock()
	trans, ok := p.Transitions[genKey(p.CurrentStatus, event)]
	if !ok {
		return "", errcode.Newf("no such transition for %s, %s", p.CurrentStatus, event)
	}

	if trans.Event != event {
		return "", errcode.New("event mismatch")
	}
	p.CurrentStatus = trans.TargetStatus

	return p.CurrentStatus, nil
}

// GetTargetTransition returns the target transition for a given status and event. If no such transition exists, it returns an error.
func (p *NamespaceTransitions) GetTargetTransition(status, event string) (*Transition, error) {
	p.rwMux.Lock()
	defer p.rwMux.Unlock()
	trans, ok := p.Transitions[genKey(status, event)]
	if !ok {
		return nil, errcode.Newf("no such transition for %s, %s", status, event)
	}
	return trans, nil
}

// SetCurrentStatus sets the current status to a new value.
func (p *NamespaceTransitions) SetCurrentStatus(status string) error {
	p.rwMux.Lock()
	defer p.rwMux.Unlock()
	for _, v := range p.Transitions {
		if v.CurrentStatus == status {
			p.CurrentStatus = status
			return nil
		}
	}
	return errcode.Newf("no such transition for status: %s", status)
}

// GetCurrentStatus returns the current status.
func (p *NamespaceTransitions) GetCurrentStatus() string {
	p.rwMux.RLock()
	defer p.rwMux.RUnlock()
	return p.CurrentStatus
}

// genKey generates a key based on the status and event.
func genKey(status, event string) string {
	return status + "::" + event
}
