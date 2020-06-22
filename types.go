// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

// Type represents the ipset type
type Type string

const (
	// HashIP represents the `hash:ip` type ipset.
	HashIP Type = "hash:ip"
)

const (
	// ProtocolFamilyIPV4 represents IPv4 protocol.
	ProtocolFamilyIPv4 = "inet"
	// ProtocolFamilyIPV4 represents IPv6 protocol.
	ProtocolFamilyIPv6 = "inet6"
)

// ValidIPSetTypes defines the supported ip set type.
var ValidIPSetTypes = []Type{
	HashIP,
}
