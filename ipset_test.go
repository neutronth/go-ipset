// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

import (
	"reflect"
	"testing"

	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

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
				{Element: "172.18.3.2", Comment: "\"ContainerID: deadbeaf\""},
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
				{Element: "172.18.3.3", Comment: "\"ContainerID: deadbeafbeaf\""},
				{Element: "172.18.3.2", Comment: "\"ContainerID: deadbeaf\""},
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
