// Copyright 2019 Dave Marsh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
)

// scan-test.go

const (
	root    = "/home/dave/scantest"
	tout    = "gen"
	tscript = "run"
)

type testBuilder struct {
}

func (b *testBuilder) Filter(info fs.DirEntry) (ok bool) {
	name := info.Name()
	if name != tout && name != tscript {
		ok = true
	}
	fmt.Println("Filter", ok)
	return
}

func (b *testBuilder) Format(info fs.DirEntry, folder *Folder) (cmd string) {
	cmd = fmt.Sprintf(`cp "%s" "%s"%s`, info.Name(), folder.Destination, "\n")
	fmt.Print("Format", cmd)
	return
}

func TestScan(t *testing.T) {
	tf, err := GenerateTestData(root)
	if err != nil {
		t.Fatal(err)
	}
	tf.Display()
	tf.Create()
	//defer tf.DestroyTestData()

	var b = &testBuilder{}
	err = os.Chdir(root)
	if err != nil {
		t.Fatal(err)
	}
	var folders Folders
	folders, err = Build(root, tout, tscript, b, true, true)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range folders {
		fmt.Println(f.Source)
	}

}
