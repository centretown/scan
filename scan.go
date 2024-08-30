// Copyright 2019 Dave Marsh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/golang-collections/collections/queue"
)

// Folder - properties collected during scan
type Folder struct {
	// Source - path of origin
	Source string
	// Destination - output path of file operations
	Destination string
	// Script - file name of stored generated text
	Script string
	// Code - generated formatted text
	Code string
	// Files - list of filtered files
	Files []fs.DirEntry
	// Children - sub folders
	Children []string
}

// Folders - collection
type Folders []*Folder

// Builder - required functions
type Builder interface {
	// Filter - function accepts or rejects file
	Filter(file fs.DirEntry) bool
	// Format - function returns formatted text for selected files
	Format(file fs.DirEntry, folder *Folder) string
}

var showMessages = false

// Build - returns input scan results in folders
// arguments:
//
//		in - input base
//		out - output base
//	 script - name of generated script
//		builder - Builder interface to filter files and format scripts
//	 write - create folders and generate scripts
//		verbose - display messages
func Build(in, out, script string, builder Builder, write bool, verbose bool) (folders Folders, err error) {
	// clear queue
	defer clearQ()

	if verbose {
		showMessages = true
		defer displayError(&err)
	}

	msg(`Build(in: "%v", write: %v, verbose: %v)`, in, write, verbose)

	err = os.Chdir(in)
	if err != nil {
		return
	}

	if !path.IsAbs(out) {
		// ensure absolute output folder
		out = path.Join(in, out)
	}

	if write {
		// create base output folder if necessary
		msg("create base output folder %v", out)
		err = makeDir(out)
		if err != nil {
			return
		}
	}

	// build folder structure from base
	out = path.Join(out, path.Base(in))

	enq(in, out)

	// scan and filter each folder in the tree
	folders, err = scanQ(script, builder)
	if err != nil {
		return
	}

	// generate scripts
	msg("generate scripts")
	folders.generate(builder)

	os.Chdir(in)
	if write {
		// create output folder
		msg("create output folder %v", out)
		err = makeDir(out)
		if err != nil {
			return
		}
		msg("write script and create output folders")
		err = folders.Write()
		if err != nil {
			return
		}
	}
	return
}

func scanQ(script string, builder Builder) (fs Folders, err error) {
	var (
		folder *Folder
		qi, qo string
	)
	// scan and filter each folder in the tree
	for fQ.Len() > 0 {
		qi, qo = deq()
		msg("scan and filter %v", qi)
		folder, err = scanFolder(qi, qo, script, builder)
		if err != nil {
			return
		}
		fs = append(fs, folder)
	}
	return
}

func scanFolder(in, out, script string, builder Builder) (folder *Folder, err error) {
	var (
		files           []fs.DirEntry
		name, inc, outc string
	)

	err = os.Chdir(in)
	if err != nil {
		return
	}

	files, err = os.ReadDir(in)
	if err != nil {
		return
	}

	folder = &Folder{Source: in, Destination: out, Script: script}
	msg("scanning: %s", in)
	for _, f := range files {
		name = f.Name()
		msg("scan:%s", name)
		if builder.Filter(f) {
			if f.IsDir() {
				inc, outc = path.Join(in, name), path.Join(out, name)
				enq(inc, outc)
				folder.Children = append(folder.Children, inc)
				msg("folder:%s selected.", name)
			} else {
				folder.Files = append(folder.Files, f)
				msg("file:%s selected.", name)
			}
		}
	}
	msg("children:%v", folder.Children)
	return
}

// format to run child scripts
var (
	formatChild = `cd "%s"` + "\n./%s\ncd ..\n"
)

// generate - create a script for the selected files and folders
func (f *Folder) generate(b Builder) {
	cmd := ""
	for _, file := range f.Files {
		cmd += b.Format(file, f)
	}

	for _, child := range f.Children {
		cmd += fmt.Sprintf(formatChild, child, f.Script)
	}

	f.Code = cmd
	msg("generate '%s/%s\n%s'", f.Source, f.Script, f.Code)
}

// Write - create folders and write output files
func (f *Folder) Write(cmd []byte) (err error) {
	msg("navigate to %s", f.Source)
	err = os.Chdir(f.Source)
	if err != nil {
		return
	}

	msg("write script '%s'", f.Script)
	err = os.WriteFile(f.Script, cmd, os.ModeAppend|os.ModePerm)
	if err != nil {
		return
	}

	msg("verify/create destination %s", f.Destination)
	err = makeDir(f.Destination)

	return
}

// Generate - a script for the selected files and folders
func (fs Folders) generate(b Builder) (cmd string) {
	for _, f := range fs {
		f.generate(b)
	}
	return
}

// Write - folders and output files
func (fs Folders) Write() (err error) {
	for _, f := range fs {
		err = f.Write([]byte(f.Code))
		if err != nil {
			return
		}
	}
	return
}

// makeDir - creates new directories if necessary
func makeDir(dir string) (err error) {
	var f os.FileInfo
	f, err = os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, os.ModeDir|os.ModePerm)
	} else if !f.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", dir)
	}
	return
}

// fQ - folders queue
var fQ = queue.New()

type qitem struct {
	in, out string
}

func enq(in, out string) {
	fQ.Enqueue(&qitem{in: in, out: out})
}

func deq() (in, out string) {
	next := fQ.Dequeue().(*qitem)
	in, out = next.in, next.out
	return
}

func clearQ() {
	for fQ.Len() > 0 {
		fQ.Dequeue()
	}
}

// msg - verbose messages
func msg(format string, a ...interface{}) {
	if showMessages {
		fmt.Printf(format+"\n", a)
	}
}

func displayError(err *error) {
	if *err != nil {
		fmt.Println(*err)
	}
}
