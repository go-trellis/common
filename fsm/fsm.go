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
)

type fsm struct {
	Transactions map[string]map[string]*Transaction

	sync.RWMutex
}

var defaultFSM *fsm

// New get default fsm
func New() Repo {
	if defaultFSM == nil {
		defaultFSM = &fsm{
			Transactions: make(map[string]map[string]*Transaction),
		}
	}
	return defaultFSM
}

// Add add a transaction
func (p *fsm) Add(t *Transaction) {
	if e := t.valid(); e != nil {
		return
	}

	p.Lock()
	defer p.Unlock()
	p.add(t)
}

func (p *fsm) add(t *Transaction) {

	spaceTrans := p.Transactions[t.Namespace]

	if spaceTrans == nil {
		spaceTrans = make(map[string]*Transaction)
	}

	spaceTrans[p.genKey(t.CurrentStatus, t.Event)] = t
	p.Transactions[t.Namespace] = spaceTrans
}

// GetTargetTransaction get trans by current information
func (p *fsm) GetTargetTransaction(namespace, curStatus, event string) *Transaction {
	p.RLock()
	defer p.RUnlock()
	return p.getTransaction(namespace, curStatus, event)
}

// Remove remove all transactions
func (p *fsm) Remove() {
	p.Lock()
	defer p.Unlock()
	p.remove()
}

func (p *fsm) remove() {
	p.Transactions = make(map[string]map[string]*Transaction)
}

// RemoveNamespace remove namespace's transactions
func (p *fsm) RemoveNamespace(namespace string) {
	if namespace == "" {
		return
	}
	p.Lock()
	defer p.Unlock()
	p.removeNamespace(namespace)
}

func (p *fsm) removeNamespace(namespace string) {
	delete(p.Transactions, namespace)
}

// RemoveByTransaction remove a transaction by current information
func (p *fsm) RemoveByTransaction(t *Transaction) {
	if e := t.validCurrent(); e != nil {
		return
	}
	p.Lock()
	defer p.Unlock()
	p.removeByTransaction(t)
}

func (p *fsm) removeByTransaction(t *Transaction) {
	spaceTrans := p.Transactions[t.Namespace]
	delete(spaceTrans, p.genKey(t.CurrentStatus, t.Event))
	p.Transactions[t.Namespace] = spaceTrans
}

func (p *fsm) getTransaction(namespace, curStatus, event string) *Transaction {
	spaceTrans := p.Transactions[namespace]
	if spaceTrans == nil {
		return nil
	}
	return spaceTrans[p.genKey(curStatus, event)]
}

func (p *fsm) genKey(curStatus, event string) string {
	return curStatus + "::" + event
}
