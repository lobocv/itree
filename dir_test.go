package main

import (
	"os"
	"testing"
	"github.com/lobocv/itree/ctx"
)

var testDirRoot = "/tmp/itree"

func setUp() error {

	var paths = []string { testDirRoot + "/a/a1/a2",
	 					   testDirRoot + "/b/b1/b2"}

	for _, p := range paths {
		err := os.MkdirAll(p, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func tearDown() {

}

func TestCreateDirectoryChain(t *testing.T) {
	var curDir *ctx.Directory

	err := setUp()
	if err != nil {
		t.Error(err)
	}
	defer tearDown()

	cwd := testDirRoot + "/a/a1/a2"

	curDir, err = ctx.CreateDirectoryChain(cwd)
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
