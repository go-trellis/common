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

// Bus xxx
type Bus interface {
	RegistEvent(eventNames ...string) error

	Subscribe(eventName string, fn func(...interface{})) (Subscriber, error)
	Unsubscribe(eventName string, ids ...string) error
	UnsubscribeAll(eventName string)

	Publish(eventName string, evt ...interface{})

	ListEvents() (events []string)
}

// DefaultEventCenterName default event center name
const DefaultEventCenterName = "trellis::event::default-center"

var defBus = NewEventCenter(DefaultEventCenterName)

// RegistEvent regist event name to bus
func RegistEvent(eventNames ...string) error {
	return defBus.RegistEvent(eventNames...)
}

// Subscribe listen to an event name and return a subscriber.
func Subscribe(eventName string, fn func(...interface{})) (Subscriber, error) {
	return defBus.Subscribe(eventName, fn)
}

// Unsubscribe cancel listen to an event name.
func Unsubscribe(eventName string, ids ...string) error {
	return defBus.Unsubscribe(eventName, ids...)
}

// Publish publish an event to bus.
func Publish(eventName string, event ...interface{}) {
	defBus.Publish(eventName, event...)
}

// ListEvents list all events in the bus.
func ListEvents() (events []string) {
	return defBus.ListEvents()
}
