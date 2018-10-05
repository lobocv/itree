package main

import (
	"fmt"
	"github.com/lobocv/itree/ctx"
	"github.com/lobocv/itree/screen"
	"github.com/nsf/termbox-go"
	"os"
	"path/filepath"
)

func Min(i, j int) int {
	if i > j {
		return j
	} else {
		return i
	}
}

func Max(i, j int) int {
	if i > j {
		return i
	} else {
		return j
	}
}

func main() {
	var inputmode = false
	var err error
	err = termbox.Init()

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

	dir := new(ctx.Directory)
	dir.SetDirectory(cwd)
	s := screen.GetScreen(dir)

MainLoop:
	for {
		s.SetDirectory(dir)
		s.Draw(4)

		ev := termbox.PollEvent()
		if inputmode {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				inputmode = false
				s.SearchString = s.SearchString[:0]
				s.SetState(screen.Directory)
			} else if ev.Key == termbox.KeyEnter {
				dir.FilterContents(string(s.SearchString))
				inputmode = false
				s.SetState(screen.Directory)
			} else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
				if len(s.SearchString) > 0 {
					s.SearchString = s.SearchString[:len(s.SearchString)-1]
				}
			} else {
				s.SearchString = append(s.SearchString, ev.Ch)
			}
			continue MainLoop
		}

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
				nextdir, err := dir.Ascend()
				if nextdir != nil && err == nil {
					dir = nextdir
				}
			case termbox.KeyArrowRight:
				nextdir, err := dir.Descend()
				if nextdir != nil && err == nil {
					dir = nextdir
				}
			case termbox.KeyCtrlH:
				if s.GetState() != screen.Help {
					s.SetState(screen.Help)
				} else {
					s.SetState(screen.Directory)
				}
			case termbox.KeyCtrlS:
				if s.GetState() != screen.Search {
					s.SetState(screen.Search)
					inputmode = true
				} else {
					s.SetState(screen.Directory)
					inputmode = false
				}
			}
		}

		switch ev.Ch {
		case 'q':
			break MainLoop
		case 'h':
			dir.SetShowHidden(!dir.ShowHidden)
		case 'a':
			dir = new(ctx.Directory)
			dir.SetDirectory(cwd)
		case 'e':
			// Move up by half the distance between the selected file
			// Always move at least 2 steps
			by := -Max(2, dir.FileIdx/2)
			dir.MoveSelector(by)
		case 'd':
			// Move down by half the distance between the selected file
			// Always move at least 2 steps
			by := Max(2, (len(dir.Files)-dir.FileIdx)/2)
			dir.MoveSelector(by)
		case 'c':
			// Toggle position between first and last file in the directory
			if dir.FileIdx == 0 {
				dir.FileIdx = len(dir.Files) - 1
			} else {
				dir.FileIdx = 0
			}

		}
	}
	// We must print the directory we end up in so that we can change to it
	fmt.Print(dir.AbsPath)
}
