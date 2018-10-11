package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"github.com/lobocv/itree/ctx"
)

var testDirRoot = "/tmp/itree"

var dirpaths = []string {testDirRoot + "/a/a1/a2",
					  	 testDirRoot + "/b/b1/b2"}

var filepaths = []string {"/a/f1", "/a/f2", "/a/f3", "/a/a1/f1", "/a/a1/a2/f2", "/a/a1/a2/f3"}


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
	for _, fp := range filepaths {
		err:= os.Remove(testDirRoot + fp)
		if err != nil {
			panic(fmt.Sprintf("Cannot remove file %v", testDirRoot + fp))
		}
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
	a.FilterContents("f")
	if len(a.FilteredFiles) != 3 {
		t.Error("Expected 3 files")
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
