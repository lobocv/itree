package main

import (
	"fmt"
	"github.com/lobocv/itree/ctx"
	"io/ioutil"
	"os"
	"testing"
)

var testDirRoot = "/tmp/itree"

var dirpaths = []string {testDirRoot + "/a/a1/a2",
					  	 testDirRoot + "/a/A1",
					  	 testDirRoot + "/b/b1/b2",
						 }

var filepaths = []string {"/a/f1", "/a/f2", "/a/f3", "/a/a1/f1", "/a/a1/.hidden", "/a/a1/a2/f2", "/a/a1/a2/f3"}


func setUp() error {
		for _, p := range dirpaths {
		err := os.MkdirAll(p, 0777)
		if err != nil {
			return err
		}
	}

	for _, fp := range filepaths {
		var contents []byte
		err := ioutil.WriteFile(testDirRoot + fp, contents, 0777)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func tearDown() {
	os.RemoveAll(testDirRoot)
}
func TestHiddenFiles(t *testing.T) {
	err := setUp()
	if err != nil {
		t.Error(err)
	}
	defer tearDown()

	curDir, err := getDirChain()
	if err != nil {
		t.Error(err)
	}
	a := curDir.Parent
	a.SetShowHidden(false)
	expected := 2
	if len(a.Files) != expected {
		t.Error(fmt.Sprintf("Expected %d files, found %d", expected, len(a.Files)))
	}
	a.SetShowHidden(true)
	expected = 3
	if len(a.Files) != expected {
		t.Error(fmt.Sprintf("Expected %d files, found %d", expected, len(a.Files)))
	}

}
func TestFilterContents(t *testing.T) {
	err := setUp()
	if err != nil {
		t.Error(err)
	}
	defer tearDown()

	curDir, err := getDirChain()
	if err != nil {
		t.Error(err)
	}
	a := curDir.Parent.Parent

	// Check filtering is working
	a.FilterContents("f")
	expected := 3
	if len(a.FilteredFiles) != expected {
		t.Error(fmt.Sprintf("Expected %d files, found %d", expected, len(a.FilteredFiles)))
	}
	// Check that filtering automatically moves index
	expected =2
	if a.FileIdx != expected {
		t.Error(fmt.Sprintf("Expected file index %d, found %d", expected, a.FileIdx))
	}
	// Check we cannot move selector down since all files before the current are not in filter
	expected = a.FileIdx
	a.MoveSelector(-1)
	if a.FileIdx != expected {
		t.Error(fmt.Sprintf("Expected file index %d, found %d", expected, a.FileIdx))
	}
	a.MoveSelector(2)
	expected += 2
	if a.FileIdx != expected {
		t.Error(fmt.Sprintf("Expected file index %d, found %d", expected, a.FileIdx))
	}
	a.MoveSelector(-2) // Undo
	expected -= 2
	if a.FileIdx != expected {
		t.Error(fmt.Sprintf("Expected file index %d, found %d", expected, a.FileIdx))
	}

	a.MoveSelector(100)
	expected = len(a.Files) -1
	if a.FileIdx != expected {
		t.Error(fmt.Sprintf("Expected file index %d, found %d", expected, a.FileIdx))
	}

	// Check clearing the filter works
	a.FilterContents("")
	expected = 0
	if len(a.FilteredFiles) != expected {
		t.Error(fmt.Sprintf("Expected %d files, found %d", expected, len(a.FilteredFiles)))
	}

}


func getDirChain() (*ctx.Directory, error) {
	cwd := testDirRoot + "/a/a1/a2"

	curDir, err := ctx.CreateDirectoryChain(cwd)
	return curDir, err
}

func TestCreateDirectoryChain(t *testing.T) {

	err := setUp()
	if err != nil {
		t.Error(err)
	}
	defer tearDown()


	curDir, err := getDirChain()
	if err != nil {
		t.Error(err)
	}

	var nodes = []string{"/tmp/itree/a/a1/a2", "/tmp/itree/a/a1", "/tmp/itree/a", "/tmp/itree", "/tmp", "/"}

	for _, p := range nodes {
		if curDir.AbsPath != p {
			t.Error("Failed to create directory chain")
		}
		curDir = curDir.Parent
	}

	if curDir != nil {
		t.Error("top level directory has a parent")
	}

}
