package main

import (
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"errors"
)

/*
Screen drawing methods
*/

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func printdircontents(dir DirContext, x, y int) error {
	files, err := ioutil.ReadDir(dir.AbsPath)
	if err != nil {
		return err
	}

	tbprint(x, y, termbox.ColorRed, termbox.ColorDefault, dir.AbsPath)
	for yoffset, f := range files {
		var color termbox.Attribute
		var itemname string

		if dir.FileIdx == yoffset {
			color =  termbox.ColorCyan
		} else {
			color =  termbox.ColorWhite
		}
		if f.IsDir() { itemname = "/"}
		itemname += f.Name()
		tbprint(x, y + yoffset+1,color, termbox.ColorDefault,  "   " + itemname)
	}

	return nil
}

/*
Directory methods
*/

type DirContext struct {
	AbsPath string
	Files []os.FileInfo
	FileIdx int
}

func (d* DirContext) SetDirectory(path string) error {
	//var err error
	if _, err := os.Stat(path); err != nil {
		return err
	}
	d.AbsPath = path
	f, err := ioutil.ReadDir(d.AbsPath)
	if err != nil {
		return err
	}
	d.Files = f
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
	f := d.Files[d.FileIdx]
	tbprint(50, 10, termbox.ColorWhite,termbox.ColorRed, f.Name())
	if f.IsDir() {
		newpath := path.Join(d.AbsPath, f.Name())
		tbprint(50, 10, termbox.ColorWhite,termbox.ColorRed, newpath)
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
	dir := DirContext{}
	dir.SetDirectory(cwd)

MainLoop:
	for {
		printdircontents(dir, 0, 1)
		termbox.Flush()

		ev := termbox.PollEvent()

		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break MainLoop

			case termbox.KeyArrowUp:
				dir.MoveSelector(-1)
				termbox.Clear(termbox.ColorDefault,termbox.ColorDefault)
			case termbox.KeyArrowDown:
				dir.MoveSelector(1)
				termbox.Clear(termbox.ColorDefault,termbox.ColorDefault)


			case termbox.KeyArrowLeft:
				dir.Ascend()
				termbox.Clear(termbox.ColorDefault,termbox.ColorDefault)
			case termbox.KeyArrowRight:
				dir.Descend()
				termbox.Clear(termbox.ColorDefault,termbox.ColorDefault)


			}

		switch ev.Ch {
		case 'q':
			break MainLoop
		}

		case termbox.EventResize:
		}
		termbox.Flush()
	}
}
