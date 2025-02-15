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

// Repo is the interface for managing namespace's transitions in cache
type Repo interface {
	// AddTransition add a transition into cache
	AddTransition(*Transition) error
	// RemoveTransition remove a transaction by information
	RemoveTransition(*Transition) error
	// ChangeCurrentStatus change namespace's current status by namespace and event
	ChangeCurrentStatus(namespace, event string) (string, error)
	// SetCurrentStatus set namespace's current status
	SetCurrentStatus(namespace, status string) error
	// GetCurrentStatus get current status
	GetCurrentStatus(namespace string) string
	// GetTargetTransition get target transition by current information
	GetTargetTransition(namespace, curStatus, event string) (*Transition, error)
	// Remove remove all namespaces from cache
	Remove()
	// AddNamespace add a namespace into cache
	AddNamespace(namespace string) error
	// RemoveNamespace remove namespace's Transitions
	RemoveNamespace(namespace string)
}

// TransitionRepo is the interface for managing transitions in cache
type TransitionRepo interface {
	// Add a transaction into cache
	AddTransaction(trans *Transition) error
	// RemoveNamespace remove namespace's Transitions
	RemoveTransition(status, event string) error
	// RemoveByTransaction remove a transaction by information
	ChangeStatus(event string) (string, error)
	// GetTargetTransition get target transition by current information
	GetTargetTransition(status, event string) (*Transition, error)
	// SetCurrentStatus set current status
	SetCurrentStatus(status string) error
	// GetCurrentStatus get current status
	GetCurrentStatus() string
}
