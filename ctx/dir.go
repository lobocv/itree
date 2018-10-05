package ctx

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

/*
Directory methods
*/

type Directory struct {
	AbsPath    string
	Files      []os.FileInfo
	FileIdx    int
	ShowHidden bool
	Parent     *Directory
}

// Methods for filtering files by directory, then file
type OSFiles []os.FileInfo

func (f OSFiles) Len() int           { return len(f) }
func (f OSFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f OSFiles) Less(i, j int) bool { return f[i].IsDir() }

func (d *Directory) SetDirectory(path string) error {
	//var err error
	if _, err := os.Stat(path); err != nil {
		return err
	}
	d.AbsPath = path
	files, err := ioutil.ReadDir(d.AbsPath)
	if err != nil {
		return err
	}

	var filtered []os.FileInfo
	// Filter out hidden files
	if !d.ShowHidden {
		filtered = files[:0]
		for _, f := range files {
			if !strings.HasPrefix(f.Name(), ".") {
				filtered = append(filtered, f)
			}
		}
		// Sort by directory
		sort.Sort(OSFiles(filtered))
	} else {
		filtered = files[:]
	}

	// Check that the index hasn't gone out of bounds
	d.Files = filtered
	if d.FileIdx > len(d.Files)-1 {
		d.FileIdx = len(d.Files) - 1
	}
	return nil
}

func (d *Directory) Ascend() (*Directory, error) {
	newpath := path.Dir(d.AbsPath)
	if d.Parent != nil {
		return d.Parent, nil
	} else {
		parent := new(Directory)

		err := parent.SetDirectory(newpath)
		if err != nil {
			return nil, err
		}

		for idx, f := range parent.Files {
			if f.Name() == newpath {
				parent.FileIdx = idx
				break
			}
		}
		return parent, err
	}

}

func (d *Directory) Descend() (*Directory, error) {
	if len(d.Files) == 0 {
		return nil, nil
	}
	child := new(Directory)
	f := d.Files[d.FileIdx]
	if f.IsDir() {
		newpath := path.Join(d.AbsPath, f.Name())
		child.SetDirectory(newpath)
		child.Parent = d
		return child, nil
	} else {
		return nil, errors.New("Cannot enter non-directory.")
	}
}

func (d *Directory) MoveSelector(dy int) {
	idx := d.FileIdx + dy
	if idx >= len(d.Files) {
		idx = len(d.Files) - 1
	} else if idx < 0 {
		idx = 0
	}
	d.FileIdx = idx
}

func (d *Directory) SetShowHidden(v bool) {
	d.ShowHidden = v
	d.SetDirectory(d.AbsPath)
}

func (d *Directory) FilterContents(searchstring string) {
	filtered := d.Files[:0]
	for _, f := range d.Files {
		if strings.Contains(f.Name(), searchstring) {
			filtered = append(filtered, f)
		}
	}
	d.Files = filtered
}
