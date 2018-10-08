package main

import (
	"fmt"
	"github.com/lobocv/itree/ctx"
	"github.com/nsf/termbox-go"
	"os"
	"path/filepath"
	"path"
	"math"
	"bytes"
	"strings"
	"errors"
	"strconv"
)

func max(i, j int) int {
	if i > j {
		return i
	} else {
		return j
	}
}

// Move up by half the distance between the selected file
// Always move at least 2 steps
func (s *Screen) JumpUp() {
	by := -max(2, s.CurrentDir.FileIdx/2)
	s.CurrentDir.MoveSelector(by)
}

// Move down by half the distance between the selected file
// Always move at least 2 steps
func (s *Screen) JumpDown() {
	by := max(2, (len(s.CurrentDir.Files)-s.CurrentDir.FileIdx)/2)
	s.CurrentDir.MoveSelector(by)
}


/*
Screen drawing methods
*/

type ScreenState int

const (
	Directory ScreenState = iota
	Help
)

type Screen struct {
	SearchString  []rune
	CurrentDir	  *ctx.Directory
	state         ScreenState
	captureInput  bool
}

func (s *Screen) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (s *Screen) PrintDirContents(x, y int, dirlist ctx.DirView) error {
	var levelOffsetX, levelOffsetY int // Draw position offset
	var stretch int                    // Length of line connecting subdirectories
	var maxLineWidth int               // Length of longest item in the directory
	var scrollOffsety int			   // Offset to scroll the visible directory text by

	screenWidth, screenHeight := termbox.Size()

	levelOffsetX = x
	levelOffsetY = y

	// Determine the scrolling offset
	scrollOffsety = levelOffsetY
	for _, dir := range dirlist {
		scrollOffsety += dir.FileIdx
	}
	scrollOffsety -= screenHeight - levelOffsetY
	if scrollOffsety < 0 {
		scrollOffsety = 0
	} else {
		pagejump := float64(screenHeight) / 5
		scrollOffsety = int(math.Ceil(float64(scrollOffsety) / pagejump) * pagejump)
	}

	// Recurse through the directory list, drawing a tree structure
	for level, dir := range dirlist {
		maxLineWidth = 0

		for ii, f := range dir.Files {

			// Keep track of the longest length item in the directory
			filename_len := len(f.Name())
			if filename_len > maxLineWidth {
				maxLineWidth = filename_len
			}

			// Determine the color of the currently printing directory item
			var color termbox.Attribute
			if dir.FileIdx == ii && level == len(dirlist)-1 {
				color = termbox.ColorCyan
			} else {
				if _, ok := dir.FilteredFiles[ii]; ok {
					color = termbox.ColorGreen
				} else if f.IsDir() {
					color = termbox.ColorYellow
				} else {
					color = termbox.ColorWhite
				}

			}

			line := bytes.Buffer{}
			if ii == 0 {
				line.WriteString(strings.Repeat("─", stretch))
			}

			switch ii {
			case 0:
				if level > 0 {
					if len(dir.Files) < 2 {
						line.WriteString("─")
					} else {
						line.WriteString("┬─")
					}
				} else {
					line.WriteString("├─")
				}
			case len(dir.Files) - 1:
				line.WriteString("└─")
			default:
				line.WriteString("├─")
			}

			// Create the item label, add / if it is a directory
			line.WriteString(f.Name())
			if f.IsDir() {
				line.WriteString("/")
			}

			// Calculate the draw position
			x := levelOffsetY + ii - scrollOffsety
			y := levelOffsetX
			if ii == 0 {
				// The first item is connected to the parent directory with a line
				// shift the position left to account for this line
				y -= stretch
			}
			if y + len(line.String()) > screenWidth && len(dirlist) > 1 {
				return errors.New("DisplayOverflow")
			}
			s.Print(y, x, color, termbox.ColorDefault, line.String())
		}

		// Determine the length of line we need to draw to connect to the next directory
		if len(dir.Files) > 0 {
			stretch = maxLineWidth - len(dir.Files[dir.FileIdx].Name())
		}

		// Shift the draw position in preparation for the next directory
		levelOffsetY += dir.FileIdx
		levelOffsetX += maxLineWidth + 2

	}

	return nil
}

func (s *Screen) ToggleHelp() ScreenState {
	if s.state != Help {
		s.state = Help
	} else {
		s.state = Directory
	}
	return s.state
}

func (s *Screen) Draw() {
	switch s.state {
	case Help:
		s.ClearScreen()
		s.Print(0, 0, termbox.ColorWhite, termbox.ColorDefault, "itree - An interactive tree application for file system navigation.")
		s.Print(0, 2, termbox.ColorWhite, termbox.ColorDefault, "h - Toggle hidden files and folders.")
		s.Print(0, 3, termbox.ColorWhite, termbox.ColorDefault, "e - Log2 skip up.")
		s.Print(0, 4, termbox.ColorWhite, termbox.ColorDefault, "d - Log2 skip down.")
		s.Print(0, 5, termbox.ColorWhite, termbox.ColorDefault, "c - Toggle position between first and last file.")
	case Directory:
		upperLevels, err := strconv.Atoi(os.Getenv("MaxUpperLevels"))
		if err != nil {
			upperLevels = 3
		}
		for {
			s.ClearScreen()
			// Print the current path
			s.Print(0, 0, termbox.ColorRed, termbox.ColorDefault, s.CurrentDir.AbsPath)
			if s.captureInput {
				instruction := "Enter a search string:"
				s.Print(0, 1, termbox.ColorWhite, termbox.ColorDefault, instruction)
				s.Print(len(instruction)+2, 1, termbox.ColorWhite, termbox.ColorDefault, string(s.SearchString))
			}
			dirlist := s.getDirView(upperLevels)
			err := s.PrintDirContents(0, 2, dirlist)
			if err == nil {
				break
			} else {
				upperLevels -= 1
			}
		}
	}

	termbox.Flush()
}

func (s *Screen) ClearScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func (s *Screen) getDirView(upperLevels int ) ctx.DirView {
	// Create a slice of the directory chain containing upperLevels number of parents
	dir := s.CurrentDir
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

func (s *Screen) Descend()  {
	dir := s.CurrentDir
	dir.Descend()
	s.SearchString = s.SearchString[:0]
	dir.FilterContents(string(s.SearchString))
	nextdir, err := dir.Descend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}

func (s *Screen) Ascend() {
	s.captureInput = false
	s.SearchString = s.SearchString[:0]
	s.CurrentDir.FilterContents(string(s.SearchString))
	nextdir, err := s.CurrentDir.Ascend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}


func (s *Screen) Main(dirpath string) string {


MainLoop:
	for {
		s.Draw()

		ev := termbox.PollEvent()
		if s.captureInput {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				s.captureInput = false
				s.SearchString = s.SearchString[:0]
				s.CurrentDir.FilterContents(string(s.SearchString))
			} else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
				if len(s.SearchString) > 0 {
					s.SearchString = s.SearchString[:len(s.SearchString)-1]
					s.CurrentDir.FilterContents(string(s.SearchString))
				}
			} else if ev.Ch != 0 {
				s.SearchString = append(s.SearchString, ev.Ch)
				s.CurrentDir.FilterContents(string(s.SearchString))
			}
			//continue MainLoop
		}

		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				break MainLoop
			case termbox.KeyArrowUp:
				s.CurrentDir.MoveSelector(-1)
			case termbox.KeyArrowDown:
				s.CurrentDir.MoveSelector(1)
			case termbox.KeyArrowLeft:
				s.Ascend()
			case termbox.KeyArrowRight:
				s.Descend()
			case termbox.KeyPgup:
				s.JumpUp()
			case termbox.KeyPgdn:
				s.JumpDown()
			case termbox.KeyCtrlH:
				s.ToggleHelp()
			}
		}

		switch ev.Ch {
		case 'q':
			break MainLoop
		case '/':
			s.captureInput = true
			s.SearchString = s.SearchString[:0]
		case 'h':
			s.CurrentDir.SetShowHidden(!s.CurrentDir.ShowHidden)
		case 'a':
			for s.CurrentDir.Parent != nil {
				s.Ascend()
			}
		case 'e':
			s.JumpUp()
		case 'd':
			s.JumpDown()
		case 'c':
			// Toggle position between first and last file in the directory
			if s.CurrentDir.FileIdx == 0 {
				s.CurrentDir.FileIdx = len(s.CurrentDir.Files) - 1
			} else {
				s.CurrentDir.FileIdx = 0
			}

		}
	}

	// Return the directory we end up in
	currentItem, err := s.CurrentDir.CurrentFile()
	if err == nil && currentItem.IsDir() && os.Getenv("EnterLastSelected") == "1" {
		return  path.Join(s.CurrentDir.AbsPath, currentItem.Name())
	} else {
		return s.CurrentDir.AbsPath
	}

}

func main() {
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		panic("Cannot get current working directory")
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		panic("Cannot get absolute directory.")
	}


	// Initialize the library that draws to the terminal
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()


	pathlist := ctx.GetPathComponents(cwd)
	var curDir, prevDir, nextDir *ctx.Directory
	for _, subdir := range pathlist {

		nextDir = new(ctx.Directory)
		nextDir.SetDirectory(subdir)
		nextDir.Parent = prevDir
		if prevDir != nil {
			prevDir.Child = nextDir
			for ii, f := range prevDir.Files {
				if strings.HasSuffix(subdir, f.Name()) {
					prevDir.FileIdx = ii
					break
				}
			}
		}
		prevDir = nextDir
	}
	// Set the current directory context
	curDir = nextDir

	s := Screen{make([]rune, 0, 100), curDir, Directory, false}
	finalPath := s.Main(cwd)
	// We must print the directory we end up in so that we can change to it
	// If we end up selecting a directory item, then change into that directory,
	// If we end up on a file item, change into that files directory

	fmt.Print(finalPath)
}
