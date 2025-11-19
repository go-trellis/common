package main

import (
	"fmt"

	"trellis.tech/trellis/common.v3/state-machine/fsm"
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
