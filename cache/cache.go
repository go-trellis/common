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

package cache

import "time"

// Cache Manager functions for executing k-v tables base on TableCache
type Cache interface {
	// All returns a list of all tables at the node.
	All() []string
	// GetTableCache get TableCache
	GetTableCache(tab string) (TableCache, bool)
	// New Creates a new table.
	New(tab string, options ...OptionFunc) error
	// Insert the object or all of the objects in list.
	Insert(tab string, key, value interface{}) bool
	// InsertExpire inserts the object or all of the objects with expired time in list.
	InsertExpire(tab string, key, value interface{}, expire time.Duration) bool
	// Delete deletes the entire table Tab.
	Delete(tab string) bool
	// DeleteObject deletes all objects with key, Key from table Tab.
	DeleteObject(tab string, key interface{}) bool
	// DeleteObjects delete all objects in the table Tab. Remain table in cache.
	DeleteObjects(tab string)
	// Lookup Look up values with key, Key from table Tab.
	Lookup(tab string, key interface{}) ([]interface{}, bool)
	// LookupAll Look up all values in the Tab.
	LookupAll(tab string) (map[interface{}][]interface{}, bool)
	// Member Returns true if one or more elements in the table has key, otherwise false.
	Member(tab string, key interface{}) bool
	// Members Returns all keys in the table Tab.
	Members(tab string) ([]interface{}, bool)
	// SetExpire Set key expire time in the table Tab.
	SetExpire(tab string, key interface{}, expire time.Duration) bool
}
