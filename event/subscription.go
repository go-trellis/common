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
	"fmt"
)

// Subscriber subscribe interface for event subscription.
type Subscriber interface {
	GetID() string
	Publish(values ...any) error
	Stop()
}

// NewDefSubscriber creates a new default subscriber.
func NewDefSubscriber(sub any) (Subscriber, error) {
	var subscriber Subscriber
	switch s := sub.(type) {
	case func(...any) error:
		subscriber = &defSubscriber{
			id: GenSubscriberID(),
			fn: s,
		}
	case Subscriber:
		subscriber = s
	default:
		return nil, fmt.Errorf("unkown subscriber type: %+v", s)
	}
	return subscriber, nil
}

// Subscriber is returned from the Subscribe function.
//
// This value and can be passed to Unsubscribe when the observer is no longer interested in receiving messages
type defSubscriber struct {
	id string
	fn func(values ...any) error
}

// GetID returns the ID of the subscriber.
func (p *defSubscriber) GetID() string {
	return p.id
}

// Publish publishes a message to the subscriber.
func (p *defSubscriber) Publish(values ...any) error {
	return p.fn(values...)
}

// Stop do nothing
func (*defSubscriber) Stop() {}
