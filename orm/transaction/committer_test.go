/*
Copyright © 2026 Henry Huang <hhh@rutcode.com>

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

package transaction

import (
	"database/sql"
	"testing"
)

type stubRepo struct{}

func (stubRepo) SetSession(any) error { return nil }

type stubEngine struct {
	beginTXCalls int
}

func (e *stubEngine) NewSession() (any, error) { return nil, nil }
func (e *stubEngine) Exec(string, ...any) (sql.Result, error) {
	return nil, nil
}
func (e *stubEngine) BeginTransaction() (Transaction, error) {
	e.beginTXCalls++
	return &stubTrans{isTX: true}, nil
}
func (e *stubEngine) BeginNonTransaction() (Transaction, error) {
	return &stubTrans{isTX: false}, nil
}
func (e *stubEngine) AddHook(any) error { return nil }
func (e *stubEngine) Close() error      { return nil }

type stubTrans struct {
	isTX        bool
	commitCalls int
	lastFn      any
}

func (t *stubTrans) Session() any        { return nil }
func (t *stubTrans) IsTransaction() bool { return t.isTX }
func (t *stubTrans) Commit(fn any, _ ...any) error {
	t.commitCalls++
	t.lastFn = fn
	if lf := GetLogicFunc(fn); lf == nil || lf.Logic == nil {
		return ErrNotFoundFunction
	}
	return nil
}

func TestCommitterRejectsMissingLogicBeforeBegin(t *testing.T) {
	engine := &stubEngine{}
	c := NewCommitter(map[string]Engine{DefaultDatabase: engine})

	err := c.TX(map[int]any{BeforeLogic: func() {}}, stubRepo{})
	if err != ErrNotFoundFunction {
		t.Fatalf("TX error = %v, want %v", err, ErrNotFoundFunction)
	}
	if engine.beginTXCalls != 0 {
		t.Fatalf("BeginTransaction called %d times, want 0", engine.beginTXCalls)
	}
}
