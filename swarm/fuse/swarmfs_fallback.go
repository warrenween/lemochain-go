// Copyright 2017 The lemochain-go Authors
// This file is part of the lemochain-go library.
//
// The lemochain-go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The lemochain-go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the lemochain-go library. If not, see <http://www.gnu.org/licenses/>.

// +build !linux,!darwin,!freebsd

package fuse

import (
	"errors"
)

var errNoFUSE = errors.New("FUSE is not supported on this platform")

func isFUSEUnsupportedError(err error) bool {
	return err == errNoFUSE
}

type MountInfo struct {
	MountPoint     string
	StartManifest  string
	LatestManifest string
}

func (self *SwarmFS) Mount(mhash, mountpoint string) (*MountInfo, error) {
	return nil, errNoFUSE
}

func (self *SwarmFS) Unmount(mountpoint string) (bool, error) {
	return false, errNoFUSE
}

func (self *SwarmFS) Listmounts() ([]*MountInfo, error) {
	return nil, errNoFUSE
}

func (self *SwarmFS) Stop() error {
	return nil
}
