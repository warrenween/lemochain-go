// Copyright 2016 The lemochain-go Authors
// This file is part of lemochain-go.
//
// lemochain-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// lemochain-go is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with lemochain-go. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/LemoFoundationLtd/lemochain-go/internal/cmdtest"
	"github.com/docker/docker/pkg/reexec"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "glemo-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type testglemo struct {
	*cmdtest.TestCmd

	// template variables for expect
	Datadir  string
	Lemobase string
}

func init() {
	// Run the app if we've been exec'd as "glemo-test" in runGlemo.
	reexec.Register("glemo-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func TestMain(m *testing.M) {
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

// spawns glemo with the given command line args. If the args don't set --datadir, the
// child g gets a temporary data directory.
func runGlemo(t *testing.T, args ...string) *testglemo {
	tt := &testglemo{}
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	for i, arg := range args {
		switch {
		case arg == "-datadir" || arg == "--datadir":
			if i < len(args)-1 {
				tt.Datadir = args[i+1]
			}
		case arg == "-lemobase" || arg == "--lemobase":
			if i < len(args)-1 {
				tt.Lemobase = args[i+1]
			}
		}
	}
	if tt.Datadir == "" {
		tt.Datadir = tmpdir(t)
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		args = append([]string{"-datadir", tt.Datadir}, args...)
		// Remove the temporary datadir if somlemoing fails below.
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "glemo". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("glemo-test", args...)

	return tt
}
