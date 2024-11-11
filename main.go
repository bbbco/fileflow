package main

import (
	// Standard imports
	"fmt"
	"log"
	"os"
	"regexp"

	// Local imports
	"fileflow/src/logutils"

	// Third party imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-autostart"
)

type Flow struct {
	Pattern     string
	Destination string
}

type Directory struct {
	Path  string
	Flows []Flow
}

type AppState struct {
	Directories []Directory
}

var logWriter *logutils.LogWriter

func main() {

	a := app.New()

	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("FileFlow Settings")
	w.Resize(fyne.NewSize(600, 400))

	if a, ok := a.(desktop.App); ok {
		fyneDesktopMenu := fyne.NewMenu("FileFlow",
			fyne.NewMenuItem("Settings", func() {
				w.Show()
			}))
		a.SetSystemTrayMenu(fyneDesktopMenu)
	}
	// Introduction tab
	introLabel := widget.NewLabel("Welcome to FileFlow!\nAutomate file movement based on patterns.")
	introContent := container.NewVBox(introLabel)

	state := AppState{}

	// Directory tab components
	dirList := widget.NewList(
		func() int { return len(state.Directories) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(state.Directories[i].Path)
		})

	// Empty state handling for directories
	emptyDirLabel := widget.NewLabel("No directories added yet.")
	dirListWrapper := container.NewVBox(emptyDirLabel)

	dirAddButton := widget.NewButton("Add Directory", func() {
		selectDirectory(w, &state)
	})

	dirRemoveButton := widget.NewButton("Remove Directory", func() {
		if len(state.Directories) > 0 {
			selectedDir := state.Directories[0] // Get the selected directory
			if len(selectedDir.Flows) > 0 {
				w.ShowAndRun() // Show confirmation dialog for flows removal
			} else {
				removeDirectory(&state, selectedDir)
			}
		}
	})

	// Check if there are directories and adjust UI
	if len(state.Directories) > 0 {
		dirListWrapper = container.NewVBox(dirAddButton, dirList, dirRemoveButton)
	} else {
		dirListWrapper = container.NewVBox(dirAddButton, emptyDirLabel)
	}

	// Flows tab components

	flowAddButton := widget.NewButton("Add Flow", func() {
		if len(state.Directories) > 0 {
			showAddFlowDialog(w, &state, &state.Directories[0]) // Assume the first directory is selected
		}
	})

	flowsList := widget.NewList(
		func() int { return len(state.Directories[0].Flows) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(state.Directories[0].Flows[i].Pattern)
		})

	// Empty state handling for flows
	emptyFlowLabel := widget.NewLabel("No flows for this directory.")
	flowsListWrapper := container.NewVBox(emptyFlowLabel)

	// Check if there are flows and adjust UI
	if len(state.Directories) > 0 && len(state.Directories[0].Flows) > 0 {
		flowsListWrapper = container.NewVBox(flowAddButton, flowsList)
	} else {
		flowsListWrapper = container.NewVBox(flowAddButton, emptyFlowLabel)
	}

	// Log Viewer Tab
	logContent := widget.NewMultiLineEntry()
	logContent.SetText("Execution Log...\n") // Replace with actual logs
	logContent.Disable()
	logWriter = logutils.NewLogWriter(logContent)

	// Settings Tab
	startupToggle := widget.NewCheck("Start FileFlow on system startup", func(checked bool) {
		executable, _ := os.Executable()
		app := &autostart.App{
			Name:        "FileFlow",
			DisplayName: "FileFlow",
			Exec:        []string{executable},
		}
		if checked {
			if err := app.Enable(); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := app.Disable(); err != nil {
				log.Fatal(err)
			}
		}
		// Logic to set startup preference
	})
	settingsContent := container.NewVBox(startupToggle)

	// Tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Introduction", introContent),
		container.NewTabItem("Directories", dirListWrapper),
		container.NewTabItem("Flows", flowsListWrapper),
		container.NewTabItem("Logs", logContent),
		container.NewTabItem("Settings", settingsContent),
	)

	w.SetContent(tabs)

	w.SetCloseIntercept(func() {
		w.Hide()
	})
	w.ShowAndRun()

}

func showAddDirectoryDialog(win fyne.Window, state *AppState) {
	dirDialog := widget.NewEntry()
	dirDialog.SetPlaceHolder("Enter directory path...")
	saveButton := widget.NewButton("Save", func() {
		dirPath := dirDialog.Text
		if dirPath == "" || !isValidDirectory(dirPath) {
			return
		}
		state.Directories = append(state.Directories, Directory{Path: dirPath})
	})
	win.SetContent(container.NewVBox(dirDialog, saveButton))
}

// Function to open the file dialog for directory selection
func selectDirectory(win fyne.Window, state *AppState) {
	dirDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil || folder == nil {
			fmt.Println("No directory selected.")
			return
		}
		// Here, the file URI will point to the directory selected
		logWriter.Write(fmt.Sprint("Directory selected:", folder.String()))
		state.Directories = append(state.Directories, Directory{Path: folder.Path()})
	}, win)

	dirDialog.Show()
}

func isValidDirectory(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func showAddFlowDialog(win fyne.Window, state *AppState, dir *Directory) {
	flowPatternEntry := widget.NewEntry()
	flowPatternEntry.SetPlaceHolder("Enter regex pattern...")
	saveButton := widget.NewButton("Save", func() {
		pattern := flowPatternEntry.Text
		if pattern == "" || !isValidRegex(pattern) {
			return
		}
		dir.Flows = append(dir.Flows, Flow{Pattern: pattern, Destination: dir.Path})
	})
	win.SetContent(container.NewVBox(flowPatternEntry, saveButton))
}

func isValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

func removeDirectory(state *AppState, dir Directory) {
	var indexToRemove int
	for i, d := range state.Directories {
		if d.Path == dir.Path {
			indexToRemove = i
			break
		}
	}
	state.Directories = append(state.Directories[:indexToRemove], state.Directories[indexToRemove+1:]...)
}
