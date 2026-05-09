# fsm
Finite-state machine in go

## Introduction

* [Click for Chinese documentation](http://zh.wikipedia.org/wiki/%E6%9C%89%E9%99%90%E7%8A%B6%E6%80%81%E6%9C%BA)
* [Click to article in English](http://en.wikipedia.org/wiki/Finite-state_machine)

## Installation

```go
go get -u github.com/go-trellis/common.v3/state-machine/fsm
```

## Usage

### fsm repo

```go
// Repo is the interface for managing namespace's transitions in cache
type Repo interface {
	// AddTransition add a transition into cache
	AddTransition(*Transition) error
	// RemoveTransition remove a transaction by information
	RemoveTransition(*Transition) error
	// ChangeCurrentStatus change namespace's current status by namespace and event
	ChangeCurrentStatus(namespace string, event string) (string, error)
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
```

### new and input a namespace's transaction

* [main.go](example/main.go)

### Sample Config

* [sample.yaml](example/sample.yaml)