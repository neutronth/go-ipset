// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

import (
	"encoding/xml"
	"fmt"

	utilexec "k8s.io/utils/exec"
)

// IPSetSet defines the XML data structure of each set.
type IPSetSet struct {
	Name string `xml:"name,attr"`
}

// IPSetSets defines the XML data structure of sets.
type IPSetSets struct {
	List []IPSetSet `xml:"ipset"`
}

// Interface is an injectable interface for running ipset commands.
// Implementations must be goroutine-safe.
type Interface interface {
	ListSets() ([]string, error)
}

// IPSetCmd represents the ipset util. We use ipset command for
// ipset execute.
const IPSetCmd = "ipset"

// IPSetCmdMandatoryArgs represents the mandatory ipset command arguments.
var IPSetCmdMandatoryArgs = []string{"-o", "xml"}

type runner struct {
	exec utilexec.Interface
}

// New returns a new Interface which will exec ipset.
func New(exec utilexec.Interface) Interface {
	return &runner{
		exec: exec,
	}
}

// cmdArgsBuilder builds the ipset command with mandatory arguments.
func cmdArgsBuilder(args []string) []string {
	return append(args, IPSetCmdMandatoryArgs...)
}

// ListSets list all set names from kernel.
func (runner *runner) ListSets() ([]string, error) {
	cmdArgs := cmdArgsBuilder([]string{"list", "-n"})
	out, err := runner.exec.
		Command(IPSetCmd, cmdArgs...).
		CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("error listing all sets, error: %v", err)
	}

	var sets IPSetSets
	err = xml.Unmarshal([]byte(out), &sets)

	if err != nil {
		return nil, fmt.Errorf("error extract data sets, error: %v", err)
	}

	list := []string{}
	for _, set := range sets.List {
		list = append(list, set.Name)
	}

	return list, nil
}
