package main

import (
	"fmt"
	"github.com/lobocv/itree/ctx"
	"github.com/lobocv/itree/screen"
	"github.com/nsf/termbox-go"
	"os"
	"path/filepath"
)

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
