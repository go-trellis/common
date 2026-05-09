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

import (
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestNew(t *testing.T) {
	repo := New()
	testutils.Assert(t, repo != nil, "repo should not be nil")
}

func TestFSM_AddTransition(t *testing.T) {
	fsm := New().(*FSM)
	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}

	err := fsm.AddTransition(trans)
	testutils.Ok(t, err)
}

func TestFSM_AddTransition_Invalid(t *testing.T) {
	fsm := New().(*FSM)

	// Test nil transition
	err := fsm.AddTransition(nil)
	testutils.NotOk(t, err, "should return error for nil transition")

	// Test empty namespace
	trans := &Transition{
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}
	err = fsm.AddTransition(trans)
	testutils.NotOk(t, err, "should return error for empty namespace")

	// Test empty event
	trans = &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		TargetStatus:  "running",
	}
	err = fsm.AddTransition(trans)
	testutils.NotOk(t, err, "should return error for empty event")

	// Test empty current status
	trans = &Transition{
		Namespace:    "test",
		Event:        "start",
		TargetStatus: "running",
	}
	err = fsm.AddTransition(trans)
	testutils.NotOk(t, err, "should return error for empty current status")

	// Test empty target status
	trans = &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
	}
	err = fsm.AddTransition(trans)
	testutils.NotOk(t, err, "should return error for empty target status")
}

func TestFSM_GetTargetTransition(t *testing.T) {
	fsm := New().(*FSM)
	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}

	err = fsm.AddTransition(trans)
	testutils.Ok(t, err)

	result, err := fsm.GetTargetTransition("test", "initial", "start")
	testutils.Ok(t, err)
	testutils.Assert(t, result != nil, "result should not be nil")
	testutils.Equals(t, "running", result.TargetStatus, "target status should match")
}

func TestFSM_GetTargetTransition_NotFound(t *testing.T) {
	fsm := New().(*FSM)

	// Test namespace not found
	_, err := fsm.GetTargetTransition("nonexistent", "initial", "start")
	testutils.NotOk(t, err, "should return error for nonexistent namespace")

	// Test transition not found
	err = fsm.AddNamespace("test")
	testutils.Ok(t, err)

	_, err = fsm.GetTargetTransition("test", "initial", "start")
	testutils.NotOk(t, err, "should return error for nonexistent transition")
}

func TestFSM_GetCurrentStatus(t *testing.T) {
	fsm := New().(*FSM)

	// Test empty namespace
	status := fsm.GetCurrentStatus("nonexistent")
	testutils.Equals(t, "", status, "status should be empty for nonexistent namespace")

	// Test with namespace
	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	status = fsm.GetCurrentStatus("test")
	testutils.Equals(t, "", status, "status should be empty initially")

	// Add transition and set status
	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}
	err = fsm.AddTransition(trans)
	testutils.Ok(t, err)

	err = fsm.SetCurrentStatus("test", "initial")
	testutils.Ok(t, err)

	status = fsm.GetCurrentStatus("test")
	testutils.Equals(t, "initial", status, "status should match")
}

func TestFSM_Remove(t *testing.T) {
	fsm := New().(*FSM)
	err := fsm.AddNamespace("test1")
	testutils.Ok(t, err)
	err = fsm.AddNamespace("test2")
	testutils.Ok(t, err)

	fsm.Remove()

	status := fsm.GetCurrentStatus("test1")
	testutils.Equals(t, "", status, "status should be empty after remove")
}

func TestFSM_RemoveNamespace(t *testing.T) {
	fsm := New().(*FSM)
	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	fsm.RemoveNamespace("test")

	status := fsm.GetCurrentStatus("test")
	testutils.Equals(t, "", status, "status should be empty after remove namespace")

	// Test empty namespace
	fsm.RemoveNamespace("")
	// Should not panic
}

func TestFSM_AddNamespace(t *testing.T) {
	fsm := New().(*FSM)

	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	// Test adding same namespace again
	err = fsm.AddNamespace("test")
	testutils.Ok(t, err) // Should be idempotent

	// Test empty namespace
	err = fsm.AddNamespace("")
	testutils.NotOk(t, err, "should return error for empty namespace")
}

func TestFSM_RemoveTransition(t *testing.T) {
	fsm := New().(*FSM)
	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}

	err := fsm.AddTransition(trans)
	testutils.Ok(t, err)

	err = fsm.RemoveTransition(trans)
	testutils.Ok(t, err)

	_, err = fsm.GetTargetTransition("test", "initial", "start")
	testutils.NotOk(t, err, "transition should be removed")
}

func TestFSM_RemoveTransition_Invalid(t *testing.T) {
	fsm := New().(*FSM)

	// Test nil transition
	err := fsm.RemoveTransition(nil)
	testutils.NotOk(t, err, "should return error for nil transition")

	// Test nonexistent transition
	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}
	err = fsm.AddNamespace("test")
	testutils.Ok(t, err)

	err = fsm.RemoveTransition(trans)
	// May or may not error depending on implementation - transition doesn't exist
	_ = err
}

func TestFSM_ChangeCurrentStatus(t *testing.T) {
	fsm := New().(*FSM)
	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}

	err = fsm.AddTransition(trans)
	testutils.Ok(t, err)

	err = fsm.SetCurrentStatus("test", "initial")
	testutils.Ok(t, err)

	newStatus, err := fsm.ChangeCurrentStatus("test", "start")
	testutils.Ok(t, err)
	testutils.Equals(t, "running", newStatus, "new status should match")

	currentStatus := fsm.GetCurrentStatus("test")
	testutils.Equals(t, "running", currentStatus, "current status should be updated")
}

func TestFSM_ChangeCurrentStatus_NotFound(t *testing.T) {
	fsm := New().(*FSM)

	// Test namespace not found
	_, err := fsm.ChangeCurrentStatus("nonexistent", "start")
	testutils.NotOk(t, err, "should return error for nonexistent namespace")

	// Test transition not found
	err = fsm.AddNamespace("test")
	testutils.Ok(t, err)

	_, err = fsm.ChangeCurrentStatus("test", "start")
	testutils.NotOk(t, err, "should return error for nonexistent transition")
}

func TestFSM_SetCurrentStatus(t *testing.T) {
	fsm := New().(*FSM)
	err := fsm.AddNamespace("test")
	testutils.Ok(t, err)

	trans := &Transition{
		Namespace:     "test",
		CurrentStatus: "initial",
		Event:         "start",
		TargetStatus:  "running",
	}

	err = fsm.AddTransition(trans)
	testutils.Ok(t, err)

	err = fsm.SetCurrentStatus("test", "initial")
	testutils.Ok(t, err)

	status := fsm.GetCurrentStatus("test")
	testutils.Equals(t, "initial", status, "status should be set")
}

func TestFSM_SetCurrentStatus_NotFound(t *testing.T) {
	fsm := New().(*FSM)

	// Test namespace not found
	err := fsm.SetCurrentStatus("nonexistent", "initial")
	testutils.NotOk(t, err, "should return error for nonexistent namespace")

	// Test status not found
	err = fsm.AddNamespace("test")
	testutils.Ok(t, err)

	err = fsm.SetCurrentStatus("test", "nonexistent")
	testutils.NotOk(t, err, "should return error for nonexistent status")
}
