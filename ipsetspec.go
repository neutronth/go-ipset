// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
//
// SPDX-License-Identifier: Apache-2.0

package ipset

type IPSetSpecFunc func(*IPSet)

// IPSetName set the name.
func IPSetName(name string) IPSetSpecFunc {
	return func(set *IPSet) {
		set.Name = name
	}
}

// IPSetType set the type.
func IPSetType(setType Type) IPSetSpecFunc {
	return func(set *IPSet) {
		set.SetType = setType
	}
}

// IPSetHashFamily set the hash family.
func IPSetHashFamily(family string) IPSetSpecFunc {
	return func(set *IPSet) {
		set.HashFamily = family
	}
}

// IPSetHashSize set the hash size.
func IPSetHashSize(size int) IPSetSpecFunc {
	return func(set *IPSet) {
		set.HashSize = size
	}
}

// IPSetMaxElement set the maximum elements that set could hold.
func IPSetMaxElement(max int) IPSetSpecFunc {
	return func(set *IPSet) {
		set.MaxElement = max
	}
}

// IPSetSpec provides the interface to setup the set specification with
// default values
func IPSetSpec(setters ...IPSetSpecFunc) *IPSet {
	set := &IPSet{
		SetType:    HashIP,
		HashFamily: ProtocolFamilyIPv4,
		HashSize:   1024,
		MaxElement: 65536,
	}

	for _, setter := range setters {
		setter(set)
	}

	return set
}
