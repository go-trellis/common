/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

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

package node

import (
	"fmt"
)

// Get value from metadata
func (p *Node) Get(key string) (interface{}, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, ok := p.Metadata[key]
	return value, ok
}

// Set kv pair from metadata
func (p *Node) Set(key string, value string) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = value
}

// Manager node manager functions defines.
type Manager interface {
	// Add adds a node to the node ring.
	Add(node *Node)
	// NodeFor get the node responsible for the data key.
	NodeFor(keys ...string) (*Node, bool)
	// Remove removes all nodes from the node ring.
	Remove()
	// RemoveByID removes a node from the node ring.
	RemoveByID(id string)
	// PrintNodes print all nodes
	PrintNodes()
	// IsEmpty is the node ring empty
	IsEmpty() bool
}

// New node manager with node type, it has no nodes
func New(nt NodeType, name string) (Manager, error) {
	switch nt {
	case NodeType_Direct:
		return NewDirect(name)
	case NodeType_Random:
		return NewRandom(name)
	case NodeType_Consistent:
		return NewConsistent(name)
	case NodeType_RoundRobin:
		return NewRoundRobin(name)
	default:
		return nil, fmt.Errorf("not supperted type: %d", nt)
	}
}

// NewWithNodes new node manager by node type with nodes
func NewWithNodes(nt NodeType, name string, nodes []*Node) (Manager, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodes should at least one")
	}

	m, err := New(nt, name)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		m.Add(n)
	}
	return m, nil
}
