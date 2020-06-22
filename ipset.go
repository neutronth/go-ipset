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

// IPSetEntry defines the XML data structure of each entry.
type IPSetEntry struct {
	Element string `xml:"elem"`
	Comment string `xml:"comment"`
}

// IPSetSet defines the XML data structure of each set.
type IPSetSet struct {
	Name    string       `xml:"name,attr"`
	Entries []IPSetEntry `xml:"members>member"`
}

// IPSetSets defines the XML data structure of sets.
type IPSetSets struct {
	List []IPSetSet `xml:"ipset"`
}

// Interface is an injectable interface for running ipset commands.
// Implementations must be goroutine-safe.
type Interface interface {
	ListSets() ([]string, error)
	ListEntries(setname string) ([]IPSetEntry, error)
	AddEntry(entry *IPSetEntry, setname string, ignoreExistErr bool) error
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

// ListSets list all set names from kernel.
func (runner *runner) ListEntries(setname string) ([]IPSetEntry, error) {
	cmdArgs := cmdArgsBuilder([]string{"list", setname})
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

	entries := []IPSetEntry{}
	for _, set := range sets.List {
		if set.Entries != nil {
			entries = set.Entries
		}
	}

	return entries, nil
}

// AddEntry adds an entry to the specified set name.
func (runner *runner) AddEntry(entry *IPSetEntry, setname string,
	ignoreExistErr bool) error {
	cmdArgs := []string{"add", setname, entry.Element}

	if len(entry.Comment) > 0 {
		cmdArgs = append(cmdArgs, "comment", entry.Comment)
	}

	if ignoreExistErr {
		cmdArgs = append(cmdArgs, "-exist")
	}

	cmdArgs = cmdArgsBuilder(cmdArgs)

	_, err := runner.exec.
		Command(IPSetCmd, cmdArgs...).
		CombinedOutput()

	if err != nil {
		return fmt.Errorf("error adding entry %+v, error: %v", entry, err)
	}

	return nil
}
