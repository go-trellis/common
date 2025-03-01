/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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

package event

import (
	"errors"
	"fmt"
	"sync"
)

// Center is a event center.
type Center struct {
	locker *sync.RWMutex
	name   string
	groups map[string]SubscriberGroup
}

// NewEventCenter creates a new event center.
func NewEventCenter(name string) Bus {
	if len(name) == 0 {
		panic(errors.New("center name is empty"))
	}
	return &Center{
		locker: &sync.RWMutex{},
		groups: make(map[string]SubscriberGroup),
	}
}

// Name returns the name of this event center.
func (p *Center) Name() string {
	return p.name
}

// RegistEvent registers events.
func (p *Center) RegistEvent(eventNames ...string) error {
	if len(eventNames) == 0 {
		return nil
	}

	p.locker.Lock()
	defer p.locker.Unlock()
	for _, eventName := range eventNames {
		if len(eventName) == 0 {
			return errors.New("center event is empty")
		}

		if _, exist := p.groups[eventName]; exist {
			return fmt.Errorf("event name [%s] is already in groups", eventName)
		}

		p.groups[eventName] = NewSubscriberGroup()
	}
	return nil
}

// Subscribe listens to events.
func (p *Center) Subscribe(eventName string, fn func(...any)) (Subscriber, error) {
	if len(eventName) == 0 {
		return nil, errors.New("event name is empty")
	}
	p.locker.RLock()
	defer p.locker.RUnlock()
	group, exist := p.groups[eventName]
	if !exist {
		return nil, fmt.Errorf("event name [%s] is not exists", eventName)
	}
	return group.Subscriber(fn)
}

// Unsubscribe unsubscribes from events.
func (p *Center) Unsubscribe(eventName string, ids ...string) error {
	if len(eventName) == 0 {
		return errors.New("event name is empty")
	}
	p.locker.Lock()
	defer p.locker.Unlock()
	group, exist := p.groups[eventName]
	if !exist {
		return fmt.Errorf("event name [%s] is not exists", eventName)
	}

	return group.RemoveSubscriber(ids...)
}

// UnsubscribeAll unsubscribes all subscribers from an event.
func (p *Center) UnsubscribeAll(eventName string) {
	p.locker.Lock()
	defer p.locker.Unlock()
	group, exist := p.groups[eventName]
	if !exist {
		return
	}
	group.ClearSubscribers()
}

// Publish publishes events to subscribers.
func (p *Center) Publish(eventName string, evts ...any) {
	if len(eventName) == 0 {
		return
	}

	p.locker.RLock()
	group, exist := p.groups[eventName]
	p.locker.RUnlock()
	if !exist {
		return
	}

	group.Publish(evts...)
}

// ListEvents lists all events.
func (p *Center) ListEvents() (events []string) {
	p.locker.RLock()
	defer p.locker.RUnlock()
	for event := range p.groups {
		events = append(events, event)
	}
	return
}
