package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/lobocv/itree/ctx"
	"github.com/nsf/termbox-go"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

type CaptureMode int

const (
	Search CaptureMode = iota
	Command
)

type ExitCommand struct {
	command string
	args    []string
}

func (cmd *ExitCommand) FullCommand() string {
	return cmd.command + " " + strings.Join(cmd.args, " ")
}

// Screen represents the application
type Screen struct {
	CurrentDir    *ctx.Directory
	state         ScreenState
	searchString  []rune
	commandString []rune
	captureInput  bool
	captureMode   CaptureMode

	highlightedColor termbox.Attribute
	filteredColor    termbox.Attribute
	directoryColor   termbox.Attribute
	fileColor        termbox.Attribute
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
	var scrollOffsety int              // Offset to scroll the visible directory text by
	var subDirSpacing = 2              // Spacing between subdirectories (on top of max item length)

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
		scrollOffsety = int(math.Ceil(float64(scrollOffsety)/pagejump) * pagejump)
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
					color = s.filteredColor
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
			if x+len(line.String()) > screenWidth && len(dirlist) > 1 {
				return errors.New("DisplayOverflow")
			}
			if y < y0 {
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
		var lc int
		var help = []string{
			"Calvin Lobo, 2018 - https://github.com/lobocv/itree",
			"",
			"An interactive tree application for file system navigation.",
			"",
			"                           CONTROLS                             ",
			"================================================================",
		}
		hotkeys := []struct{ hotkey, description string }{
			{"Left / Right", "Enter / exit currently selected directory."},
			{"Up / Down", "Move directory item selector position by one."},
			{"ESC or q", "Exit and change directory."},
			{"CTRL + C", "Exit without changing directory."},
			{"CTRL + h", "Opens help menu to show the list of hotkey mappings."},
			{"h", "Toggle on / off visibility of hidden files."},
			{"e", "Move selector half the distance between the current position and the top of the directory."},
			{"d", "Move selector half the distance between the current position and the bottom of the directory."},
			{"c", "Toggle position."},
			{"a", "Jump up two directories."},
			{"/", "Enters input capture mode for directory filtering."},
			{":", "Enters input capture mode for exit command."},
		}
		s.clearScreen()
		for _, line := range help {
			s.Print(0, lc, termbox.ColorWhite, termbox.ColorDefault, line)
			lc++
		}
		lc++
		for _, hotkey := range hotkeys {
			hk := fmt.Sprintf("%-12v -  ", hotkey.hotkey)
			s.Print(0, lc, termbox.ColorWhite, termbox.ColorDefault, hk)
			s.Print(len(hk), lc, termbox.ColorWhite, termbox.ColorDefault, hotkey.description)
			lc++
		}
		lc += 2
		s.Print(0, lc, termbox.ColorWhite, termbox.ColorDefault, "Press q to exit this menu.")

	case Directory:
		upperLevels, err := strconv.Atoi(os.Getenv("MaxUpperLevels"))
		if err != nil {
			upperLevels = 3
		}
		for {
			s.clearScreen()
			var instruction string
			// Print the current path
			s.Print(0, 0, termbox.ColorRed, termbox.ColorDefault, s.CurrentDir.AbsPath)
			if s.captureInput {
				switch s.captureMode {
				case Search:
					instruction = "Enter a search string:  " + string(s.searchString)
				case Command:
					instruction = "Enter a terminal command and hit enter:  " + string(s.commandString)
				}
				s.Print(0, 1, termbox.ColorWhite, termbox.ColorDefault, instruction)
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
func (s *Screen) getDirView(upperLevels int) ctx.DirView {
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
func (s *Screen) enterCurrentDirectory() {
	dir := s.CurrentDir
	dir.Descend()
	s.searchString = s.searchString[:0]
	dir.FilterContents(string(s.searchString))
	nextdir, err := dir.Descend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}

// Exits the current directory.
func (s *Screen) exitCurrentDirectory() {
	s.captureInput = false
	s.searchString = s.searchString[:0]
	s.CurrentDir.FilterContents(string(s.searchString))
	nextdir, err := s.CurrentDir.Ascend()
	if nextdir != nil && err == nil {
		s.CurrentDir = nextdir
	}
}

func (s *Screen) setCaptureMode(mode CaptureMode) {
	s.captureMode = mode

}

// Sets the application in the mode to capture input for the search string
func (s *Screen) startCapturingInput() {
	s.captureInput = true
	switch s.captureMode {
	case Search:
		s.searchString = s.searchString[:]
	case Command:
		s.commandString = s.commandString[:]
	}

}

// Exits the mode to capture input
func (s *Screen) stopCapturingInput() {
	s.captureInput = false
	if s.captureMode == Search {
		s.searchString = s.searchString[:0]
		s.CurrentDir.FilterContents(string(s.searchString))
	}
}

// Add a character to the currently capturing string
func (s *Screen) appendToCaptureInput(ch rune) {
	switch s.captureMode {
	case Search:
		s.searchString = append(s.searchString, ch)
		s.CurrentDir.FilterContents(string(s.searchString))
	case Command:
		s.commandString = append(s.commandString, ch)
	}
}

// Remove a character from the currently capturing string
func (s *Screen) popFromCaptureInput() {
	switch s.captureMode {
	case Search:
		if len(s.searchString) > 0 {
			s.searchString = s.searchString[:len(s.searchString)-1]
			s.CurrentDir.FilterContents(string(s.searchString))
		}
	case Command:
		if len(s.commandString) > 0 {
			s.commandString = s.commandString[:len(s.commandString)-1]
		}
	}

}

// Toggle position between first and last file in the directory
func (s *Screen) toggleIndexToExtremities() {
	if s.CurrentDir.FileIdx == 0 {
		s.CurrentDir.FileIdx = len(s.CurrentDir.Files) - 1
	} else {
		s.CurrentDir.FileIdx = 0
	}
}

// Main loop of the application
func (s *Screen) Main() ExitCommand {

MainLoop:
	for {
		s.draw()

		ev := termbox.PollEvent()
		if s.captureInput {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				s.stopCapturingInput()
				continue
			} else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
				s.popFromCaptureInput()
			} else if ev.Ch != 0 || ev.Key == termbox.KeySpace {
				s.appendToCaptureInput(ev.Ch)
				continue MainLoop
			}
		}

		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				if s.state == Help {
					s.toggleHelp()
				} else {
					break MainLoop
				}
			case termbox.KeyCtrlC:
				return ExitCommand{"", nil}
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
			case termbox.KeyEnter:
				if s.captureInput && s.captureMode == Command {
					if curFile, err := s.CurrentDir.CurrentFile(); err == nil {
						return ExitCommand{command: string(s.commandString),
							args: []string{path.Join(s.CurrentDir.AbsPath, curFile.Name())}}
					}
				}

			}
			switch ev.Ch {
			case 'q':
				if s.state == Help {
					s.toggleHelp()
				} else {
					break MainLoop
				}
			case '/':
				s.setCaptureMode(Search)
				s.startCapturingInput()
			case ':':
				s.setCaptureMode(Command)
				s.startCapturingInput()
			case 'h':
				s.CurrentDir.SetShowHidden(!s.CurrentDir.ShowHidden)
			case 'a':
				s.exitCurrentDirectory()
				s.exitCurrentDirectory()
			case 'e':
				s.jumpUp()
			case 'd':
				s.jumpDown()
			case 'c':
				s.toggleIndexToExtremities()
			}
		}

	}

	// Return the directory we end up in
	currentItem, err := s.CurrentDir.CurrentFile()
	if err == nil && currentItem.IsDir() && os.Getenv("EnterLastSelected") == "1" {
		return ExitCommand{command: "cd", args: []string{path.Join(s.CurrentDir.AbsPath, currentItem.Name())}}
	} else {
		return ExitCommand{command: "cd", args: []string{s.CurrentDir.AbsPath}}
	}

}

// Print error message to stderr and exit with error code 1
func fatal(err error) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%v", err))
	os.Exit(1)
}

func main() {
	var err error

	for _, arg := range os.Args {
		switch arg {
		case "-h", "--help":
			fmt.Fprintln(os.Stderr, "itree - A visual file system navigation tool.\n"+
				"Press h for information on hotkeys.")
			os.Exit(0)
		}
	}

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

	// Set the current directory context
	var curDir *ctx.Directory
	curDir, err = ctx.CreateDirectoryChain(cwd)
	if curDir == nil {
		fatal(err)
	}
	if err != nil {
		fatal(err)
	}

	s := Screen{searchString: make([]rune, 0, 100),
		commandString:    make([]rune, 0, 100),
		CurrentDir:       curDir,
		state:            Directory,
		captureMode:      Search,
		highlightedColor: termbox.ColorCyan,
		filteredColor:    termbox.ColorGreen,
		directoryColor:   termbox.ColorYellow,
		fileColor:        termbox.ColorWhite,
	}
	exitCommand := s.Main()
	// Print the command we want to execute in the current shell
	// The companion shell script will execute this command in the current shell.
	fmt.Print(exitCommand.FullCommand())
}
