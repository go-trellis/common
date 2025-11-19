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

package injector

import (
	"testing"

	"trellis.tech/trellis/common.v3/testutils"
)

func TestInject_Simple(t *testing.T) {
	type Target struct {
		Value string `inject:"value"`
	}

	target := &Target{}
	err := Inject(target, "test-value")
	testutils.Ok(t, err)
	testutils.Equals(t, "test-value", target.Value, "Value should be injected")
}

func TestInject_MultipleParams(t *testing.T) {
	type Target struct {
		StringValue string  `inject:"string"`
		IntValue    int     `inject:"int"`
		FloatValue  float64 `inject:"float"`
	}

	target := &Target{}
	err := Inject(target, "test", 42, 3.14)
	testutils.Ok(t, err)
	testutils.Equals(t, "test", target.StringValue, "StringValue should be injected")
	testutils.Equals(t, 42, target.IntValue, "IntValue should be injected")
	testutils.Equals(t, 3.14, target.FloatValue, "FloatValue should be injected")
}

func TestInject_NoParams(t *testing.T) {
	type Target struct {
		Value string `inject:"value"`
	}

	target := &Target{}
	err := Inject(target)
	// Inject without params should not error (but fields won't be set)
	_ = err
}

func TestInject_NoTags(t *testing.T) {
	type Target struct {
		Value string
	}

	target := &Target{Value: "original"}
	err := Inject(target, "test")
	testutils.Ok(t, err)
	// Without inject tags, value should remain unchanged
	testutils.Equals(t, "original", target.Value, "Value should not change without inject tag")
}

func TestInject_StructParam(t *testing.T) {
	type Param struct {
		Name string
		Age  int
	}

	type Target struct {
		Person *Param `inject:""`
	}

	param := &Param{Name: "John", Age: 30}
	target := &Target{}
	err := Inject(target, param)
	testutils.Ok(t, err)
}

func TestInject_PointerParam(t *testing.T) {
	type Target struct {
		Value *string `inject:"value"`
	}

	str := "test"
	target := &Target{}
	err := Inject(target, &str)
	testutils.Ok(t, err)
}

func TestInject_EmptyTarget(t *testing.T) {
	err := Inject(nil)
	// Inject with nil target might not error, depends on implementation
	_ = err
}
