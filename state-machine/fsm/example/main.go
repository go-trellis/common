/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

package main

import (
	"fmt"

	"github.com/go-trellis/common.v3/state-machine/fsm"
)

func main() {
	repo, err := fsm.NewFSMRepoFromConfigFile("./sample.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(repo)

	fmt.Println(repo.GetTargetTransition("namespace1", "status1", "event1"))
	fmt.Println(repo.ChangeCurrentStatus("namespace1", "event1"))
	fmt.Println(repo.GetCurrentStatus("namespace1"))

	fmt.Println(repo.AddTransition(&fsm.Transition{
		Namespace: "namespace1", CurrentStatus: "status11", Event: "failed", TargetStatus: "status111"}))

	fmt.Println(repo.ChangeCurrentStatus("namespace1", "failed"))
	fmt.Println(repo.GetCurrentStatus("namespace1"))

	fmt.Println(repo.ChangeCurrentStatus("namespace2", "event2"))

	fmt.Println(repo.ChangeCurrentStatus("namespace2", "event41"))

	repo.RemoveNamespace("namespace2")
	fmt.Println(repo.GetCurrentStatus("namespace2"))

	repo.Remove()
	fmt.Println(repo.GetCurrentStatus("namespace1"))
}
