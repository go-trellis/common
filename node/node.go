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

	"trellis.tech/trellis/common.v0/config"
)

// Type define node type
type Type uint8

// NodeType
const (
	NodeTypeDirect Type = iota
	NodeTypeRandom
	NodeTypeConsistent
	NodeTypeRoundRobin
)

// Node params for a node
type Node struct {
	// for recognize node with input id
	ID string `yaml:"id" json:"id"`
	// node's probability weight, roundrobin does not support
	Weight uint32 `yaml:"weight" json:"weight"`
	// node's value
	Value string `yaml:"value" json:"value"`
	// kvs for meta data
	Metadata config.Options `yaml:"options" json:"options"`

	number uint32
}

// Get value from metadata
func (p *Node) Get(key string) (interface{}, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, ok := p.Metadata[key]
	return value, ok
}

// Set kv pair from metadata
func (p *Node) Set(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = config.Options{}
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

// New new node manager by node type, it has no nodes
func New(nt Type, name string) (Manager, error) {
	switch nt {
	case NodeTypeDirect:
		return NewDirect(name)
	case NodeTypeRandom:
		return NewRadmon(name)
	case NodeTypeConsistent:
		return NewConsistent(name)
	case NodeTypeRoundRobin:
		return NewRoundRobin(name)
	default:
		return nil, fmt.Errorf("not supperted type: %d", nt)
	}
}

// NewWithNodes new node manager by node type with nodes
func NewWithNodes(nt Type, name string, nodes []*Node) (Manager, error) {
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
