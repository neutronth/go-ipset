// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

func TestIPSetSpec(t *testing.T) {
	cases := []struct {
		name          string
		set           *IPSet
		expectedError error
	}{
		{
			name: "Set with default specification",
			set: IPSetSpec(
				IPSetName("foo"),
			),
			expectedError: nil,
		},
		{
			name: "Set with type specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
			),
			expectedError: nil,
		},
		{
			name: "Set with type, hashsize, max elements specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
				IPSetHashSize(256),
				IPSetMaxElement(128),
			),
			expectedError: nil,
		},
		{
			name: "Set with invalid type specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType("invalid"),
			),
			expectedError: fmt.Errorf("invalid Set Type"),
		},
		{
			name: "Set with invalid hash size specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
				IPSetHashSize(-1),
			),
			expectedError: fmt.Errorf("invalid Hash Size value -1, should be >0"),
		},
		{
			name: "Set with invalid max elements specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
				IPSetHashSize(1024),
				IPSetMaxElement(0),
			),
			expectedError: fmt.Errorf("invalid Max Element value 0, should be >0"),
		},
	}

	for _, c := range cases {
		err := c.set.Validate()
		if err != c.expectedError && err.Error() != c.expectedError.Error() {
			t.Errorf("expected error: %v, got: %v", c.expectedError, err)
		}
	}
}

func TestCreateSet(t *testing.T) {
	cases := []struct {
		name              string
		set               *IPSet
		combinedOutputLog [][]string
	}{
		{
			name: "Create set foo hash:ip with default specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
			),
			combinedOutputLog: [][]string{
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "1024", "maxelem", "65536",
					"-o", "xml"},
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "1024", "maxelem", "65536",
					"-exist", "-o", "xml"},
			},
		},
		{
			name: "Create set foo hash:ip with custom specification",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
				IPSetHashSize(256),
				IPSetMaxElement(128),
			),
			combinedOutputLog: [][]string{
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "256", "maxelem", "128",
					"-o", "xml"},
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "256", "maxelem", "128",
					"-exist", "-o", "xml"},
			},
		},
		{
			name: "Create set foo hash:ip with comment option",
			set: IPSetSpec(
				IPSetName("foo"),
				IPSetType(HashIP),
				IPSetHashSize(256),
				IPSetMaxElement(128),
				IPSetWithComment(),
			),
			combinedOutputLog: [][]string{
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "256", "maxelem", "128", "comment",
					"-o", "xml"},
				{"ipset", "create", "foo", string(HashIP), "family", "inet",
					"hashsize", "256", "maxelem", "128", "comment",
					"-exist", "-o", "xml"},
			},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Failure
				func() ([]byte, []byte, error) {
					return []byte("ipset v7.6: Set cannot be created: set with the same name already exists"), nil, &fakeexec.FakeExitError{Status: 1}
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		err := runner.CreateSet(c.set, false)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}

		err = runner.CreateSet(c.set, true)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 2 {
			t.Errorf("[%s] expected 2 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[1]...).
			HasAll(c.combinedOutputLog[1]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[1])
		}

		err = runner.CreateSet(c.set, false)
		if err == nil {
			t.Errorf("[%s] expected failure, got: nil", c.name)
		}
	}
}

func TestDestroySet(t *testing.T) {
	cases := []struct {
		name              string
		setname           string
		combinedOutputLog [][]string
	}{
		{
			name:    "Destroy foo set",
			setname: "foo",
			combinedOutputLog: [][]string{
				{"ipset", "destroy", "foo", "-o", "xml"},
			},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Failure
				func() ([]byte, []byte, error) {
					return []byte("ipset v7.6: The set with the given name does not exist"), nil, &fakeexec.FakeExitError{Status: 1}
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		err := runner.DestroySet(c.setname)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}

		err = runner.DestroySet(c.setname)
		if err == nil {
			t.Errorf("[%s] expected failure, got: nil", c.name)
		}

		if fcmd.CombinedOutputCalls != 2 {
			t.Errorf("[%s] expected 2 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}
	}
}

func TestListSets(t *testing.T) {
	cases := []struct {
		name     string
		output   []byte
		expected []string
	}{
		{
			name: "1 set",
			output: []byte(`
			<ipsets>
				<ipset name="foo"/>
			</ipsets>
			`),
			expected: []string{"foo"},
		},
		{
			name: "2 sets",
			output: []byte(`
			<ipsets>
				<ipset name="foo"/>
				<ipset name="bar"/>
			</ipsets>
			`),
			expected: []string{"foo", "bar"},
		},
		{
			name: "3 sets",
			output: []byte(`
			<ipsets>
				<ipset name="foo"/>
				<ipset name="bar"/>
				<ipset name="baz"/>
			</ipsets>
			`),
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "empty sets",
			output:   []byte(`<ipsets></ipsets>`),
			expected: []string{},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) {
					return []byte(c.output), nil, nil
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		list, err := runner.ListSets()
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if len(list) != len(c.expected) {
			t.Errorf("[%s] expected %d sets, got: %d", c.name, len(c.expected),
				len(list))
		}

		if !reflect.DeepEqual(list, c.expected) {
			t.Errorf("[%s] expected sets: %v, got: %v", c.name, c.expected, list)
		}
	}
}

func TestListEntries(t *testing.T) {
	cases := []struct {
		name     string
		setname  string
		output   []byte
		expected []IPSetEntry
	}{
		{
			name:    "foo set",
			setname: "foo",
			output: []byte(`
			<ipsets>
				<ipset name="foo">
					<type>hash:ip</type>
					<revision>4</revision>
					<header>
						<family>inet</family>
						<hashsize>1024</hashsize>
						<maxelem>65536</maxelem>
						<comment/>
						<memsize>334</memsize>
						<references>0</references>
						<numentries>0</numentries>
					</header>
					<members>
						<member>
							<elem>172.18.3.2</elem>
							<comment>"ContainerID: deadbeaf"</comment>
						</member>
					</members>
				</ipset>
			</ipsets>
			`),
			expected: []IPSetEntry{
				{Element: "172.18.3.2", Comment: "ContainerID: deadbeaf"},
			},
		},
		{
			name:    "foo set, 2 entries",
			setname: "foo",
			output: []byte(`
			<ipsets>
				<ipset name="foo">
					<type>hash:ip</type>
					<revision>4</revision>
					<header>
						<family>inet</family>
						<hashsize>1024</hashsize>
						<maxelem>65536</maxelem>
						<comment/>
						<memsize>472</memsize>
						<references>0</references>
						<numentries>0</numentries>
					</header>
					<members>
						<member>
							<elem>172.18.3.3</elem>
							<comment>"ContainerID: deadbeafbeaf"</comment>
						</member>
						<member>
							<elem>172.18.3.2</elem>
							<comment>"ContainerID: deadbeaf"</comment>
						</member>
					</members>
				</ipset>
			</ipsets>
			`),
			expected: []IPSetEntry{
				{Element: "172.18.3.3", Comment: "ContainerID: deadbeafbeaf"},
				{Element: "172.18.3.2", Comment: "ContainerID: deadbeaf"},
			},
		},
		{
			name:    "foo set empty",
			setname: "foo",
			output: []byte(`
			<ipsets>
				<ipset name="foo">
					<type>hash:ip</type>
					<revision>4</revision>
					<header>
						<family>inet</family>
						<hashsize>1024</hashsize>
						<maxelem>65536</maxelem>
						<comment/>
						<memsize>200</memsize>
						<references>0</references>
						<numentries>0</numentries>
					</header>
					<members>
					</members>
				</ipset>
			</ipsets>
			`),
			expected: []IPSetEntry{},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) {
					return []byte(c.output), nil, nil
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		list, err := runner.ListEntries(c.setname)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !reflect.DeepEqual(list, c.expected) {
			t.Errorf("[%s] expected sets: %v, got: %v", c.name, c.expected, list)
		}
	}
}

func TestAddEntry(t *testing.T) {
	cases := []struct {
		name              string
		setname           string
		entry             IPSetEntry
		combinedOutputLog [][]string
	}{
		{
			name:    "Add entry",
			setname: "foo",
			entry: IPSetEntry{
				Element: "172.18.3.2",
				Comment: "ContainerID: deadbeaf",
			},
			combinedOutputLog: [][]string{
				{
					"ipset", "add", "foo", "172.18.3.2",
					"comment", "ContainerID: deadbeaf",
					"-o", "xml",
				},
				{
					"ipset", "add", "foo", "172.18.3.2",
					"comment", "ContainerID: deadbeaf", "-exist",
					"-o", "xml",
				},
			},
		},
		{
			name:    "Add entry without comment",
			setname: "bar",
			entry: IPSetEntry{
				Element: "172.18.3.2",
			},
			combinedOutputLog: [][]string{
				{"ipset", "add", "bar", "172.18.3.2", "-o", "xml"},
				{"ipset", "add", "bar", "172.18.3.2", "-exist", "-o", "xml"},
			},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Failure
				func() ([]byte, []byte, error) {
					return []byte("ipset v7.6: Element cannot be added to the set: it's already added"), nil, &fakeexec.FakeExitError{Status: 1}
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		err := runner.AddEntry(&c.entry, c.setname, false)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}

		err = runner.AddEntry(&c.entry, c.setname, true)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 2 {
			t.Errorf("[%s] expected 2 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[1]...).
			HasAll(c.combinedOutputLog[1]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[1])
		}

		err = runner.AddEntry(&c.entry, c.setname, false)
		if err == nil {
			t.Errorf("[%s] expected failure, got: nil", c.name)
		}
	}
}

func TestDelEntry(t *testing.T) {
	cases := []struct {
		name              string
		setname           string
		entryElement      string
		combinedOutputLog [][]string
	}{
		{
			name:         "Delete entry",
			setname:      "foo",
			entryElement: "172.18.3.2",
			combinedOutputLog: [][]string{
				{"ipset", "del", "foo", "172.18.3.2", "-o", "xml"},
			},
		},
	}

	for _, c := range cases {
		fcmd := fakeexec.FakeCmd{
			CombinedOutputScript: []fakeexec.FakeAction{
				// Success
				func() ([]byte, []byte, error) { return []byte{}, nil, nil },
				// Failure
				func() ([]byte, []byte, error) {
					return []byte("ipset v7.6: Element cannot be deleted from the set: it's not added"), nil, &fakeexec.FakeExitError{Status: 1}
				},
			},
		}

		fexec := fakeexec.FakeExec{
			CommandScript: []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
				func(cmd string, args ...string) exec.Cmd {
					return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
				},
			},
		}

		runner := New(&fexec)

		err := runner.DelEntry(c.entryElement, c.setname)
		if err != nil {
			t.Errorf("[%s] expected success, got: %v", c.name, err)
		}

		if fcmd.CombinedOutputCalls != 1 {
			t.Errorf("[%s] expected 1 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}

		err = runner.DelEntry(c.entryElement, c.setname)
		if err == nil {
			t.Errorf("[%s] expected failure, got: nil", c.name)
		}

		if fcmd.CombinedOutputCalls != 2 {
			t.Errorf("[%s] expected 2 CombinedOutput() calls, got: %d",
				c.name, fcmd.CombinedOutputCalls)
		}

		if !sets.NewString(fcmd.CombinedOutputLog[0]...).
			HasAll(c.combinedOutputLog[0]...) {
			t.Errorf("wrong CombinedOutput() log, got: %s",
				fcmd.CombinedOutputLog[0])
		}
	}
}
