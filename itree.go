package main

import (
	"fmt"
	"github.com/lobocv/itree/ctx"
	"github.com/lobocv/itree/screen"
	"github.com/nsf/termbox-go"
	"os"
	"path/filepath"
	"path"
)

func Max(i, j int) int {
	if i > j {
		return i
	} else {
		return j
	}
}

// Move up by half the distance between the selected file
// Always move at least 2 steps
func JumpUp(dir *ctx.Directory) {
	by := -Max(2, dir.FileIdx/2)
	dir.MoveSelector(by)
}

// Move down by half the distance between the selected file
// Always move at least 2 steps
func JumpDown(dir *ctx.Directory) {
	by := Max(2, (len(dir.Files)-dir.FileIdx)/2)
	dir.MoveSelector(by)
}


func getDirView(dir *ctx.Directory, upperLevels int ) ctx.DirView {

	// Create a slice of the directory chain containing upperLevels number of parents
	dirlist := make([]*ctx.Directory, 0, 1+upperLevels)
	dirlist = append(dirlist, dir)
	next := dir.Parent
	for ii := 0; next != nil; ii++ {
		if ii >= upperLevels {
			break
		}
		dirlist = append([]*ctx.Directory{next}, dirlist...)
		next = next.Parent
	}
	return dirlist
}

func main() {
	var inputmode = false
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		panic("Cannot get current working directory")
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		panic("Cannot get absolute directory.")
	}

	pathlist := ctx.GetPathComponents(cwd)
	var curDir, prevDir, nextDir *ctx.Directory
	for _, subdir := range pathlist {

		nextDir = new(ctx.Directory)
		nextDir.SetDirectory(subdir)
		nextDir.Parent = prevDir
		if prevDir != nil {
			prevDir.Child = nextDir
		}
		prevDir = nextDir
	}
	// Set the current directory context
	curDir = nextDir

	// Initialize the library that draws to the terminal
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	s := screen.GetScreen()

MainLoop:
	for {
		// A portion of the full path that we can draw
		dirView := getDirView(curDir, 3)
		s.Draw(dirView)

		ev := termbox.PollEvent()
		if inputmode {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				inputmode = false
				s.SearchString = s.SearchString[:0]
				s.SetState(screen.Directory)
			} else if ev.Key == termbox.KeyEnter {
				curDir.FilterContents(string(s.SearchString))
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
				curDir.MoveSelector(-1)
			case termbox.KeyArrowDown:
				curDir.MoveSelector(1)
			case termbox.KeyArrowLeft:
				nextdir, err := curDir.Ascend()
				if nextdir != nil && err == nil {
					curDir = nextdir
				}
			case termbox.KeyArrowRight:
				nextdir, err := curDir.Descend()
				if nextdir != nil && err == nil {
					curDir = nextdir
				}
			case termbox.KeyPgup:
				JumpUp(curDir)
			case termbox.KeyPgdn:
				JumpDown(curDir)
			case termbox.KeyCtrlH:
				if s.GetState() != screen.Help {
					s.SetState(screen.Help)
				} else {
					s.SetState(screen.Directory)
				}
			}
		}

		switch ev.Ch {
		case 'q':
			break MainLoop
		case '/':
			if s.GetState() != screen.Search {
				s.SetState(screen.Search)
				inputmode = true
			} else {
				s.SetState(screen.Directory)
				inputmode = false
			}
		case 'h':
			curDir.SetShowHidden(!curDir.ShowHidden)
		case 'a':
			curDir = new(ctx.Directory)
			curDir.SetDirectory(cwd)
		case 'e':
			JumpUp(curDir)
		case 'd':
			JumpDown(curDir)
		case 'c':
			// Toggle position between first and last file in the directory
			if curDir.FileIdx == 0 {
				curDir.FileIdx = len(curDir.Files) - 1
			} else {
				curDir.FileIdx = 0
			}

		}
	}
	// We must print the directory we end up in so that we can change to it
	// If we end up selecting a directory item, then change into that directory,
	// If we end up on a file item, change into that files directory
	var finalPath string
	//curDir = dirView[len(dirView)]
	currentItem, err := curDir.CurrentFile()
	if err == nil && currentItem.IsDir() && os.Getenv("EnterLastSelected") == "1" {
		finalPath = path.Join(curDir.AbsPath, currentItem.Name())
	} else {
		finalPath = curDir.AbsPath
	}
	fmt.Print(finalPath)
}
