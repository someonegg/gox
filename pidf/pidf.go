// Copyright 2015 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pidf can generate pid file.
package pidf

import (
	"io/ioutil"
	"os"
	"strconv"
)

type PidFile struct {
	path string
	pid  string
}

func NewPidFile(path string) *PidFile {
	p := &PidFile{
		path: path,
		pid:  strconv.Itoa(os.Getpid()),
	}

	ioutil.WriteFile(p.path, []byte(p.pid), 0644)

	return p
}

func (p *PidFile) Close() error {
	if pid, err := ioutil.ReadFile(p.path); err == nil {
		if len(pid) > 0 && string(pid) != p.pid {
			return nil // not me
		}
	}
	return os.Remove(p.path)
}
