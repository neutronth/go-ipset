// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	ipset "github.com/neutronth/go-ipset"
	utilexec "k8s.io/utils/exec"
)

func main() {
	var setname = "foo"
	runner := ipset.New(utilexec.New())

	set := ipset.IPSetSpec(
		ipset.IPSetName(setname),
		ipset.IPSetType(ipset.HashIP),
		ipset.IPSetWithComment(),
	)

	err := runner.CreateSet(set, true)
	if err != nil {
		fmt.Printf("Could not create set %v: error %v", set, err)
		os.Exit(1)
	}

	fmt.Println("Create Set: OK")

	err = runner.AddEntry(&ipset.IPSetEntry{
		Element: "172.18.3.2",
		Comment: "ContainerID: deadbeaf",
	}, setname, true)
	if err != nil {
		fmt.Printf("Could not add entry, error %v", err)
		os.Exit(1)
	}
	fmt.Println("Add Entry to Set: OK")

	_, err = runner.ListEntries(setname)
	if err != nil {
		fmt.Printf("Could not list entries, error %v", err)
		os.Exit(1)
	}
	fmt.Println("List entries: OK")

	err = runner.DelEntry("172.18.3.2", setname)
	if err != nil {
		fmt.Printf("Could not delete entry, error %v", err)
		os.Exit(1)
	}
	fmt.Println("Delete Entry from Set: OK")

	err = runner.DestroySet(setname)
	if err != nil {
		fmt.Printf("Could not destroy set, error %v", err)
		os.Exit(1)
	}
	fmt.Println("Destroy Set: OK")
}
