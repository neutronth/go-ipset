// Copyright 2020 Neutron Soutmun <neutron@neutron.in.th>
// Copyright 2017 The Kubernetes Authors.
//
// SPDX-License-Identifier: Apache-2.0

package ipset

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
	"k8s.io/apimachinery/pkg/util/wait"
)

type locker struct {
	lockfilePath string
	lock         *os.File
}

func (l *locker) Lock() error {
	var err error
	var success bool

	defer func(l *locker) {
		if !success {
			// Clean up immediately on failure
			l.Unlock()
		}
	}(l)

	l.lock, err = os.OpenFile(l.lockfilePath, os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open ipset lock %s: %v", l.lockfilePath, err)
	}

	err = wait.PollImmediate(200*time.Millisecond, 2*time.Second,
		func() (bool, error) {
			err := grabIPSetFileLock(l.lock)
			if err != nil {
				return false, nil
			}
			return true, nil
		})

	if err != nil {
		return fmt.Errorf("failed to acquire ipset lock: %v", err)
	}

	success = true
	return nil
}

func (l *locker) Unlock() {
	if l.lock != nil {
		l.lock.Close()
	}
}

func grabIPSetFileLock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}
