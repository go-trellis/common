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

// Repo the functions of fsm interface
type Repo interface {
	// Add a transaction into cache
	Add(*Transaction)
	// Remove all transactions
	Remove()
	// RemoveNamespace remove namespace's transactions
	RemoveNamespace(namespace string)
	// RemoveByTransaction remove a transaction by information
	RemoveByTransaction(*Transaction)
	// GetTargetTransaction get target transaction by current information
	GetTargetTransaction(namespace, curStatus, event string) *Transaction
}
