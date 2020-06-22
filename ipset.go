// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

import (
	"encoding/xml"
	"fmt"
	"strconv"

	utilexec "k8s.io/utils/exec"
)

// IPSetEntry defines the XML data structure of each entry.
type IPSetEntry struct {
	Element string `xml:"elem"`
	Comment string `xml:"comment"`
}

// IPSet defines the XML data structure of each set.
type IPSet struct {
	Name       string       `xml:"name,attr"`
	SetType    Type         `xml:"type"`
	HashFamily string       `xml:"header>family"`
	HashSize   int          `xml:"header>hashsize"`
	MaxElement int          `xml:"header>maxelem"`
	Entries    []IPSetEntry `xml:"members>member"`
}

// Validate checks if a given ipset is valid or not.
func (set *IPSet) Validate() error {
	if set.SetType == HashIP {
		if !set.validateHashFamily() {
			return fmt.Errorf("invalid Hash Family")
		}
	}

	if !set.validateIPSetType() {
		return fmt.Errorf("invalid Set Type")
	}

	if set.HashSize <= 0 {
		return fmt.Errorf("invalid Hash Size value %d, should be >0",
			set.HashSize)
	}

	if set.MaxElement <= 0 {
		return fmt.Errorf("invalid Max Element value %d, should be >0",
			set.MaxElement)
	}

	return nil
}

// checks if given set type is valid
func (set *IPSet) validateIPSetType() bool {
	for _, valid := range ValidIPSetTypes {
		if set.SetType == valid {
			return true
		}
	}

	return false
}

// checks if given hash family is supported in ipset
func (set *IPSet) validateHashFamily() bool {
	if set.HashFamily == ProtocolFamilyIPv4 ||
		set.HashFamily == ProtocolFamilyIPv6 {
		return true
	}

	return false
}

// IPSets defines the XML data structure of sets.
type IPSets struct {
	List []IPSet `xml:"ipset"`
}

// Interface is an injectable interface for running ipset commands.
// Implementations must be goroutine-safe.
type Interface interface {
	CreateSet(set *IPSet, ignoreExistErr bool) error
	DestroySet(setname string) error
	ListSets() ([]string, error)
	ListEntries(setname string) ([]IPSetEntry, error)
	AddEntry(entry *IPSetEntry, setname string, ignoreExistErr bool) error
	DelEntry(entryElement string, setname string) error
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

// CreateSet creates a new set with provided specification.
func (runner *runner) CreateSet(set *IPSet, ignoreExistErr bool) error {
	err := set.Validate()
	if err != nil {
		return fmt.Errorf("error creating set: %v, error: %v", set, err)
	}

	return runner.createSet(set, ignoreExistErr)
}

// createSet implements the create new set with validated specification
func (runner *runner) createSet(set *IPSet, ignoreExistErr bool) error {
	cmdArgs := []string{"create", set.Name, string(set.SetType)}

	if set.SetType == HashIP {
		cmdArgs = append(cmdArgs,
			"family", set.HashFamily,
			"hashsize", strconv.Itoa(set.HashSize),
			"maxelem", strconv.Itoa(set.MaxElement),
		)
	}

	if ignoreExistErr {
		cmdArgs = append(cmdArgs, "-exist")
	}

	cmdArgs = cmdArgsBuilder(cmdArgs)
	_, err := runner.exec.
		Command(IPSetCmd, cmdArgs...).
		CombinedOutput()

	if err != nil {
		return fmt.Errorf("error creating set: %v, error: %v", set, err)
	}

	return nil
}

// DestroySet destroys the specified set name.
func (runner *runner) DestroySet(setname string) error {
	cmdArgs := cmdArgsBuilder([]string{"destroy", setname})
	_, err := runner.exec.
		Command(IPSetCmd, cmdArgs...).
		CombinedOutput()

	if err != nil {
		return fmt.Errorf("error destroying set %s, error: %v", setname, err)
	}

	return nil
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

	var sets IPSets
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

	var sets IPSets
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

// DelEntry deletes an entry from the specified set name.
func (runner *runner) DelEntry(entryElement string, setname string) error {
	cmdArgs := cmdArgsBuilder([]string{"del", setname, entryElement})
	_, err := runner.exec.
		Command(IPSetCmd, cmdArgs...).
		CombinedOutput()

	if err != nil {
		return fmt.Errorf("error deleting entry %s, error: %v",
			entryElement, err)
	}

	return nil
}
