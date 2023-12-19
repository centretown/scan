// Copyright 2019 Dave Marsh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan

// data_test.go - generates, clean test data file structure

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/golang-collections/collections/queue"
)

// TestFolder - test folders and files
type TestFolder struct {
	name     string
	files    []string
	children []*TestFolder
}

//	root    = ".test"

var (
	movies  = "movies"
	movie   = "movie %02d"
	tvshows = "tv"
	series  = "Series %02d"
	season  = "Season %02d"
	episode = "%s.s%02de%02d%s"
	music   = "music"
	artist  = "Artist %02d"
	album   = "Album %02d"
	title   = "%02d - Title %d%s"
	vext    = []string{".avi", ".mp4", ".mkv", ".mpeg", ".mpg", ".wmv"}
	aext    = []string{".mp3", ".flac", ".wav"}
)

// GenerateTestData - test folders and files
func GenerateTestData(root string) (r *TestFolder, err error) {
	var cur string
	if !path.IsAbs(root) {
		cur, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		root = path.Join(cur, root)
	}

	r = &TestFolder{name: root}
	mv := buildMovies(r)
	tv := buildTv(r)
	mu := buildMusic(r)
	r.children = append(r.children, mv, tv, mu)

	//create(r)
	//r.display()
	return
}

func buildMovies(r *TestFolder) (s *TestFolder) {
	s = &TestFolder{name: path.Join(r.name, movies)}
	count := len(vext)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf(movie, i+1)
		a := &TestFolder{name: path.Join(s.name, name)}
		a.files = append(a.files, name+vext[i])
		s.children = append(s.children, a)
	}
	return
}

func buildTv(r *TestFolder) (s *TestFolder) {
	s = &TestFolder{name: path.Join(r.name, tvshows)}
	count := len(vext)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf(series, i+1)
		a := &TestFolder{name: path.Join(s.name, name)}
		for j := 0; j < count; j++ {
			bname := fmt.Sprintf(season, j+1)
			b := &TestFolder{name: path.Join(a.name, bname)}
			for k := 0; k < count; k++ {
				fname := fmt.Sprintf(episode, name, j+1, k+1, vext[j])
				b.files = append(b.files, fname)
			}
			a.children = append(a.children, b)
		}
		s.children = append(s.children, a)
	}
	return
}

func buildMusic(r *TestFolder) (s *TestFolder) {
	s = &TestFolder{name: path.Join(r.name, music)}
	count := len(aext)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf(artist, i+1)
		a := TestFolder{name: path.Join(s.name, name)}
		for j := 0; j < count; j++ {
			bname := fmt.Sprintf(album, j+1)
			b := TestFolder{name: path.Join(a.name, bname)}
			for k := 0; k < count; k++ {
				fname := fmt.Sprintf(title, k+1, k+1, aext[j])
				b.files = append(b.files, fname)
			}
			a.children = append(a.children, &b)
		}
		s.children = append(s.children, &a)
	}
	return
}

// Display -
func (tf *TestFolder) Display() {
	fmt.Println(tf.name)
	for _, f := range tf.files {
		fmt.Println("\t" + f)
	}
	for _, c := range tf.children {
		c.Display()
	}
}

// DestroyTestData -
func (tf *TestFolder) DestroyTestData() (err error) {
	err = os.RemoveAll(tf.name)
	return
}

// Create - test folders and files
func (tf *TestFolder) Create() (err error) {
	q := queue.New()
	q.Enqueue(tf)
	for q.Len() > 0 {
		qtf := q.Dequeue().(*TestFolder)

		if err = makeDir(qtf.name); err != nil {
			return
		}

		if err = os.Chdir(qtf.name); err != nil {
			return
		}

		for _, f := range qtf.files {
			os.WriteFile(f, []byte(f), os.ModePerm)
		}

		for _, c := range qtf.children {
			q.Enqueue(c)
		}
	}
	return
}
