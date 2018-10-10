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

// Create an enumeration for tracking what the screen's "state" is
// This governs what the screen should draw when .draw() is called.
type ScreenState int
const (
	Directory ScreenState = iota
	Help
)

// Screen represents the application
type Screen struct {
	SearchString  []rune
	CurrentDir	  *ctx.Directory
	state         ScreenState
	captureInput  bool

	highlightedColor termbox.Attribute
	filteredColor	 termbox.Attribute
	directoryColor	 termbox.Attribute
	fileColor	 	 termbox.Attribute
}

// Move up by half the distance between the selected file
// Always move at least 2 steps
func (s *Screen) jumpUp() {
	by := -max(2, s.CurrentDir.FileIdx/2)
	s.CurrentDir.MoveSelector(by)
}

// Move down by half the distance between the selected file
// Always move at least 2 steps
func (s *Screen) jumpDown() {
	by := max(2, (len(s.CurrentDir.Files)-s.CurrentDir.FileIdx)/2)
	s.CurrentDir.MoveSelector(by)
}

// Prints text to the terminal at the provided position and color
func (s *Screen) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

// Prints the structure of the directory path provided
func (s *Screen) drawDirContents(x0, y0 int, dirlist ctx.DirView) error {
	var levelOffsetX, levelOffsetY int // draw position offset
	var stretch int                    // Length of line connecting subdirectories
	var maxLineWidth int               // Length of longest item in the directory
	var scrollOffsety int			   // Offset to scroll the visible directory text by
	var subDirSpacing = 2			   // Spacing between subdirectories (on top of max item length)

	screenWidth, screenHeight := termbox.Size()

	levelOffsetX = x0
	levelOffsetY = y0

	// Determine the scrolling offset
	scrollOffsety = levelOffsetY
	for _, dir := range dirlist {
		scrollOffsety += dir.FileIdx
	}
	// If the selected item is off the screen then shift the entire view up in order
	// to make it visible.
	scrollOffsety -= screenHeight - levelOffsetY
	if scrollOffsety < 0 {
		scrollOffsety = 0
	} else {
		pagejump := float64(screenHeight) / 5
		scrollOffsety = int(math.Ceil(float64(scrollOffsety) / pagejump) * pagejump)
	}

	// Iterate through the directory list, drawing a tree structure
	for level, dir := range dirlist {
		maxLineWidth = 0

		for ii, f := range dir.Files {

			// Keep track of the longest length item in the directory
			filenameLen := len(f.Name())
			if filenameLen > maxLineWidth {
				maxLineWidth = filenameLen
			}

			// Determine the color of the currently printing directory item
			var color termbox.Attribute
			if dir.FileIdx == ii && level == len(dirlist)-1 {
				color = s.highlightedColor
			} else {
				if _, ok := dir.FilteredFiles[ii]; ok {
					color = s.highlightedColor
				} else if f.IsDir() {
					color = s.directoryColor
				} else {
					color = s.fileColor
				}

			}

			// Start creating the line to be printed
			line := bytes.Buffer{}
			if ii == 0 {
				line.WriteString(strings.Repeat("─", stretch))
			}

			switch ii {
			case 0:
				if level > 0 {
					if len(dir.Files) < 2 {
						line.WriteString(strings.Repeat("─", subDirSpacing))
					} else {
						line.WriteString(strings.Repeat("─", subDirSpacing))
						line.WriteString("┬─")
					}
				} else {
					line.WriteString(strings.Repeat(" ", subDirSpacing))
					line.WriteString("├─")
				}
			case len(dir.Files) - 1:
				line.WriteString(strings.Repeat(" ", subDirSpacing))
				line.WriteString("└─")
			default:
				line.WriteString(strings.Repeat(" ", subDirSpacing))
				line.WriteString("├─")
			}

			// Create the item label, add / if it is a directory
			line.WriteString(f.Name())
			if f.IsDir() {
				line.WriteString("/")
			}

			// Calculate the draw position
			y := levelOffsetY + ii - scrollOffsety
			x := levelOffsetX
			if ii == 0 {
				// The first item is connected to the parent directory with a line
				// shift the position left to account for this line
				x -= stretch
			}
			if x + len(line.String()) > screenWidth && len(dirlist) > 1 {
				return errors.New("DisplayOverflow")
			}
			if y < y0  {
				y = y0
			}
			s.Print(x, y, color, termbox.ColorDefault, line.String())
		}

		// Determine the length of line we need to draw to connect to the next directory
		if len(dir.Files) > 0 {
			stretch = maxLineWidth - len(dir.Files[dir.FileIdx].Name())
		}

		// Shift the draw position in preparation for the next directory
		levelOffsetY += dir.FileIdx
		levelOffsetX += maxLineWidth + 2 + subDirSpacing

	}

	return nil
}

// Toggles the state of the screen between regular view and the help screen
func (s *Screen) toggleHelp() ScreenState {
	if s.state != Help {
		s.state = Help
	} else {
		s.state = Directory
	}
	return s.state
}

// Draw the current representation of the screen
func (s *Screen) draw() {
	switch s.state {
	case Help:
		s.clearScreen()
		s.Print(0, 0, termbox.ColorWhite, termbox.ColorDefault, "itree - An interactive tree application for file system navigation.")
		s.Print(0, 2, termbox.ColorWhite, termbox.ColorDefault, "Calvin Lobo, 2018")
		s.Print(0, 3, termbox.ColorWhite, termbox.ColorDefault, "https://github.com/lobocv/itree")
		s.Print(0, 5, termbox.ColorWhite, termbox.ColorDefault, "Usage:")
		s.Print(0, 7, termbox.ColorWhite, termbox.ColorDefault, "h - Toggle hidden files and folders.")
		s.Print(0, 8, termbox.ColorWhite, termbox.ColorDefault, "e - Log2 skip up.")
		s.Print(0, 9, termbox.ColorWhite, termbox.ColorDefault, "d - Log2 skip down.")
		s.Print(0, 10, termbox.ColorWhite, termbox.ColorDefault, "c - Toggle position between first and last file.")
		s.Print(0, 11, termbox.ColorWhite, termbox.ColorDefault, "/ - Goes into input mode for file searching. Press ESC / CTRL+C to exit input mode.")
	case Directory:
		upperLevels, err := strconv.Atoi(os.Getenv("MaxUpperLevels"))
		if err != nil {
			upperLevels = 3
		}
		for {
			s.clearScreen()
			// Print the current path
			s.Print(0, 0, termbox.ColorRed, termbox.ColorDefault, s.CurrentDir.AbsPath)
			if s.captureInput {
				instruction := "Enter a search string:"
				s.Print(0, 1, termbox.ColorWhite, termbox.ColorDefault, instruction)
				s.Print(len(instruction)+2, 1, termbox.ColorWhite, termbox.ColorDefault, string(s.SearchString))
			}
			dirlist := s.getDirView(upperLevels)
			err := s.drawDirContents(0, 2, dirlist)
			if err == nil {
				break
			} else {
				upperLevels -= 1
			}
		}
	}

	termbox.Flush()
}

// Clear the contents of the screen
func (s *Screen) clearScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

// Get a subset of the directory chain as a slice where the last element is the current directory
// upperLevels is the number of directory levels above the current directory to include in the slice.
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

// Enters the currently selected directory
func (s *Screen) enterCurrentDirectory()  {
	dir := s.CurrentDir
	dir.Descend()
	s.SearchString = s.SearchString[:0]
	dir.FilterContents(string(s.SearchString))
	nextdir, err := dir.Descend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}

// Exits the current directory.
func (s *Screen) exitCurrentDirectory() {
	s.captureInput = false
	s.SearchString = s.SearchString[:0]
	s.CurrentDir.FilterContents(string(s.SearchString))
	nextdir, err := s.CurrentDir.Ascend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}


// Main loop of the application
func (s *Screen) Main(dirpath string) string {

	MainLoop:
	for {
		s.draw()

		ev := termbox.PollEvent()
		if s.captureInput {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				s.captureInput = false
				s.SearchString = s.SearchString[:0]
				s.CurrentDir.FilterContents(string(s.SearchString))
				continue
			} else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
				if len(s.SearchString) > 0 {
					s.SearchString = s.SearchString[:len(s.SearchString)-1]
					s.CurrentDir.FilterContents(string(s.SearchString))
				}
			} else if ev.Ch != 0 {
				s.SearchString = append(s.SearchString, ev.Ch)
				s.CurrentDir.FilterContents(string(s.SearchString))
				continue MainLoop
			}
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
				s.exitCurrentDirectory()
			case termbox.KeyArrowRight:
				s.enterCurrentDirectory()
			case termbox.KeyPgup:
				s.jumpUp()
			case termbox.KeyPgdn:
				s.jumpDown()
			case termbox.KeyCtrlH:
				s.toggleHelp()
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
					s.exitCurrentDirectory()
				}
			case 'e':
				s.jumpUp()
			case 'd':
				s.jumpDown()
			case 'c':
				// Toggle position between first and last file in the directory
				if s.CurrentDir.FileIdx == 0 {
					s.CurrentDir.FileIdx = len(s.CurrentDir.Files) - 1
				} else {
					s.CurrentDir.FileIdx = 0
				}
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

	s := Screen{make([]rune, 0, 100),
				curDir,
				Directory,
				false,
				termbox.ColorCyan,
				termbox.ColorGreen,
				termbox.ColorYellow,
				termbox.ColorWhite,

				}
	finalPath := s.Main(cwd)
	// We must print the directory we end up in so that we can change to it
	// If we end up selecting a directory item, then change into that directory,
	// If we end up on a file item, change into that files directory
	fmt.Print(finalPath)
}
