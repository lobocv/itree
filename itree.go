package main

import (
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"errors"
	"fmt"
	"strings"
	"sort"
)

/*
Screen drawing methods
*/


type Screen struct {
	Width, Height int
}

func GetScreen() Screen {
	w, h := termbox.Size()
	return Screen{w, h}
}

func (s* Screen) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (s* Screen) PrintDirContents(dir DirContext, x, y int) error {
	width, height := termbox.Size()
	width = width
	s.Print(x, y, termbox.ColorRed, termbox.ColorDefault, dir.AbsPath)
	for yoffset, f := range dir.Files {
		var color termbox.Attribute
		var itemname string

		if dir.FileIdx == yoffset {
			color =  termbox.ColorCyan
		} else {
			if f.IsDir() {
				color = termbox.ColorYellow
			} else {
				color =  termbox.ColorWhite
			}

		}
		if f.IsDir() { itemname = "/"}
		itemname += f.Name()
		row := (y + yoffset+1) % height
		col := int((y + yoffset+1) / height)
		if row < height {
			s.Print(col * 20, row, color, termbox.ColorDefault, "   " + itemname)
		}
	}

	return nil
}

func (d* Screen) ClearScreen() {
	termbox.Clear(termbox.ColorDefault,termbox.ColorDefault)
}

/*
Directory methods
*/

type DirContext struct {
	AbsPath string
	Files []os.FileInfo
	FileIdx int
	ShowHidden bool
}

// Methods for filtering files by directory, then file
type OSFiles []os.FileInfo
func (f OSFiles) Len() int { return len(f)}
func (f OSFiles) Swap(i, j int) { f[i], f[j] = f[j], f[i]}
func (f OSFiles) Less(i, j int) bool { return f[i].IsDir() }

func (d* DirContext) SetDirectory(path string) error {
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
	if ! d.ShowHidden {
		filtered = files[:0]
		for _, f := range files{
			if ! strings.HasPrefix(f.Name(), ".") {
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
		d.FileIdx = len(d.Files)-1
	}
	return nil
}

func (d* DirContext) Ascend() error {
	newpath := path.Dir(d.AbsPath)
	err := d.SetDirectory(newpath)
	for idx, f := range d.Files {
		if f.Name() == newpath {
			d.FileIdx = idx
			break
		}
	}
	return err
}

func (d* DirContext) Descend() error {
	if len(d.Files) == 0 {
		return nil
	}
	f := d.Files[d.FileIdx]
	if f.IsDir() {
		newpath := path.Join(d.AbsPath, f.Name())
		d.SetDirectory(newpath)

		return nil
	} else {
		return errors.New("Cannot enter non-directory.")
	}
}


func (d* DirContext) MoveSelector(dy int) {
	idx := d.FileIdx + dy
	if idx >= len(d.Files) {
		idx = len(d.Files) -1
	} else if idx < 0 {
		idx = 0
	}
	d.FileIdx = idx
}

func (d* DirContext) SetShowHidden(v bool) {
	d.ShowHidden = v
	d.SetDirectory(d.AbsPath)
}

/*
Application
*/

func main() {
	err := termbox.Init()

	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	cwd, err := os.Getwd()
	if err != nil {
		panic("Cannot get current working directory")
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		panic("Cannot get absolute directory.")
	}

	screen := GetScreen()

	dir := DirContext{}
	dir.SetDirectory(cwd)

MainLoop:
	for {
		screen.ClearScreen()
		screen.PrintDirContents(dir, 0, 1)
		termbox.Flush()

		ev := termbox.PollEvent()

		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break MainLoop

			case termbox.KeyArrowUp:
				dir.MoveSelector(-1)
			case termbox.KeyArrowDown:
				dir.MoveSelector(1)
			case termbox.KeyArrowLeft:
				dir.Ascend()
			case termbox.KeyArrowRight:
				dir.Descend()
			}

		switch ev.Ch {
		case 'q':
			break MainLoop
		case 'h':
			dir.SetShowHidden(!dir.ShowHidden)

		}

		case termbox.EventResize:
		}
	}

	// We must print the directory we end up in so that we can change to it
	fmt.Print(dir.AbsPath)
}
