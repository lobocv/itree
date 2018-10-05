package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

/*
Screen drawing methods
*/

type ScreenState int

const (
	Directory ScreenState = iota
	Help
	Search
)

type Screen struct {
	dir           *DirContext
	Width, Height int
	SearchString  []rune
	state         ScreenState
}

func GetScreen(dir *DirContext) Screen {
	w, h := termbox.Size()
	return Screen{dir, w, h, make([]rune, 0, 100), Directory}
}

func (s *Screen) SetDirectory(dir *DirContext) {
	s.dir = dir
}

func (s *Screen) Print(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func (s *Screen) PrintDirContents(upperLevels int) error {
	var levelOffsetX, levelOffsetY int // Draw position offset
	var stretch int                    // Length of line connecting subdirectories
	var maxLineWidth int               // Length of longest item in the directory

	// Print the current path
	s.Print(levelOffsetX, levelOffsetY, termbox.ColorRed, termbox.ColorDefault, s.dir.AbsPath)
	levelOffsetY += 2

	// Create a slice of the directory chain containing upperLevels number of parents
	dirlist := make([]*DirContext, 0, 1+upperLevels)
	next := s.dir.Parent
	for ii := 0; next != nil; ii++ {
		dirlist = append([]*DirContext{next}, dirlist...)
		next = next.Parent
		if ii >= upperLevels {
			break
		}
	}
	dirlist = append(dirlist, s.dir)

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
			if dir.FileIdx == ii {
				color = termbox.ColorCyan
			} else {
				if f.IsDir() {
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
			x := levelOffsetY + ii
			y := levelOffsetX
			if ii == 0 {
				// The first item is connected to the parent directory with a line
				// shift the position left to account for this line
				y -= stretch
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

func (s *Screen) GetState() ScreenState {
	return s.state
}
func (s *Screen) SetState(state ScreenState) {
	s.state = state
}

func (s *Screen) Draw(upperLevels int) {
	s.ClearScreen()
	switch s.state {
	case Help:
		s.Print(0, 0, termbox.ColorWhite, termbox.ColorDefault, "itree - An interactive tree application for file system navigation.")
		s.Print(0, 2, termbox.ColorWhite, termbox.ColorDefault, "h - Toggle hidden files and folders.")
	case Directory:
		s.PrintDirContents(upperLevels)
	case Search:
		s.Print(0, 0, termbox.ColorWhite, termbox.ColorDefault, "Enter a search string:")
		s.Print(0, 2, termbox.ColorWhite, termbox.ColorDefault, string(s.SearchString))
	}

	termbox.Flush()
}

func (d *Screen) ClearScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

/*
Directory methods
*/

type DirContext struct {
	AbsPath    string
	Files      []os.FileInfo
	FileIdx    int
	ShowHidden bool
	Parent     *DirContext
}

// Methods for filtering files by directory, then file
type OSFiles []os.FileInfo

func (f OSFiles) Len() int           { return len(f) }
func (f OSFiles) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f OSFiles) Less(i, j int) bool { return f[i].IsDir() }

func (d *DirContext) SetDirectory(path string) error {
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

func (d *DirContext) Ascend() (*DirContext, error) {
	newpath := path.Dir(d.AbsPath)
	if d.Parent != nil {
		return d.Parent, nil
	} else {
		parent := new(DirContext)

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

func (d *DirContext) Descend() (*DirContext, error) {
	if len(d.Files) == 0 {
		return nil, nil
	}
	child := new(DirContext)
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

func (d *DirContext) MoveSelector(dy int) {
	idx := d.FileIdx + dy
	if idx >= len(d.Files) {
		idx = len(d.Files) - 1
	} else if idx < 0 {
		idx = 0
	}
	d.FileIdx = idx
}

func (d *DirContext) SetShowHidden(v bool) {
	d.ShowHidden = v
	d.SetDirectory(d.AbsPath)
}

func (d *DirContext) FilterContents(searchstring string) {
	filtered := d.Files[:0]
	for _, f := range d.Files {
		if strings.Contains(f.Name(), searchstring) {
			filtered = append(filtered, f)
		}
	}
	d.Files = filtered
}

/*
Application
*/

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

	dir := new(DirContext)
	dir.SetDirectory(cwd)
	screen := GetScreen(dir)

MainLoop:
	for {
		screen.SetDirectory(dir)
		screen.Draw(2)

		ev := termbox.PollEvent()
		if inputmode {
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				inputmode = false
				screen.SearchString = screen.SearchString[:0]
				screen.SetState(Directory)
			} else if ev.Key == termbox.KeyEnter {
				dir.FilterContents(string(screen.SearchString))
				inputmode = false
				screen.SetState(Directory)
			} else if ev.Key == termbox.KeyBackspace2 || ev.Key == termbox.KeyBackspace {
				if len(screen.SearchString) > 0 {
					screen.SearchString = screen.SearchString[:len(screen.SearchString)-1]
				}
			} else {
				screen.SearchString = append(screen.SearchString, ev.Ch)
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
				if screen.GetState() != Help {
					screen.SetState(Help)
				} else {
					screen.SetState(Directory)
				}
			case termbox.KeyCtrlS:
				if screen.GetState() != Search {
					screen.SetState(Search)
					inputmode = true
				} else {
					screen.SetState(Directory)
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
