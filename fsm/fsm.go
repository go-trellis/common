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

	"trellis.tech/trellis/common.v3/errcode"
)

type FSM struct {
	transitions map[string]*NamespaceTransitions

	mu *sync.RWMutex
}

// Repo interface for Transition repository
func New() Repo {
	return &FSM{
		transitions: make(map[string]*NamespaceTransitions),
		mu:          &sync.RWMutex{},
	}
}

// AddTransition add transition to fsm
func (p *FSM) AddTransition(t *Transition) error {
	if err := t.valid(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.addTransition(t)
}

// addTransition add transition to fsm without lock protection. it's used for internal use only.
func (p *FSM) addTransition(trans *Transition) error {
	spaceTrans := p.transitions[trans.Namespace]

	if spaceTrans == nil {
		spaceTrans = &NamespaceTransitions{
			Namespace:   trans.Namespace,
			Transitions: map[string]*Transition{},
		}
	}
	return spaceTrans.AddTransaction(trans)
}

// GetTargetTransition get trans by current information
func (p *FSM) GetTargetTransition(namespace, status, event string) (*Transition, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	spaceTrans := p.transitions[namespace]
	if spaceTrans == nil {
		return nil, errcode.Newf("namespace transition not found")
	}
	return spaceTrans.GetTargetTransition(status, event)
}

// GetCurrentStatus get current status transition by namespace. it's used for internal use only.
func (p *FSM) GetCurrentStatus(namespace string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	spaceTrans := p.transitions[namespace]
	if spaceTrans == nil {
		return ""
	}
	return spaceTrans.GetCurrentStatus()
}

// Remove remove all NamespaceTransitions
func (p *FSM) Remove() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.transitions = map[string]*NamespaceTransitions{}
}

// RemoveNamespace remove namespace's Transitions
func (p *FSM) RemoveNamespace(namespace string) {
	if namespace == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.transitions, namespace)
}

func (p *FSM) AddNamespace(namespace string) error {
	if namespace == "" {
		return errcode.Newf("namespace cannot be empty")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.transitions[namespace]; !ok {
		p.transitions[namespace] = &NamespaceTransitions{
			Namespace:   namespace,
			Transitions: map[string]*Transition{},
		}
	}
	return nil
}

// RemoveTransition remove a Transition by current information
func (p *FSM) RemoveTransition(t *Transition) error {
	if err := t.valid(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	spaceTrans := p.transitions[t.Namespace]
	if spaceTrans == nil {
		return nil
	}
	return spaceTrans.RemoveTransition(t.CurrentStatus, t.Event)
}

func (p *FSM) ChangeCurrentStatus(namespace string, event string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	spaceTrans := p.transitions[namespace]
	if spaceTrans == nil {
		return "", errcode.Newf("namespace %s not found", namespace)
	}
	return spaceTrans.ChangeStatus(event)
}

func (p *FSM) SetCurrentStatus(namespace, status string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	spaceTrans := p.transitions[namespace]
	if spaceTrans == nil {
		return errcode.Newf("namespace %s not found", namespace)
	}
	return spaceTrans.SetCurrentStatus(status)
}
