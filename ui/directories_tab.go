package ui

import (
	"fileflow/appstate"
	"fileflow/src/logutils"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DirectoriesTab struct {
	preferences fyne.Preferences
	state       *appstate.AppState
	logWriter   *logutils.LogWriter
	content     *fyne.Container
	window      fyne.Window
}

func NewDirectoriesTab(preferences fyne.Preferences, state *appstate.AppState, window fyne.Window, logWriter *logutils.LogWriter) *DirectoriesTab {
	content := container.NewVBox()
	return &DirectoriesTab{preferences: preferences, state: state, window: window, logWriter: logWriter, content: content}
}

func (dt DirectoriesTab) GetContent() *fyne.Container {
	dt.state.LoadFromPreferences(dt.preferences)
	dt.UpdateDisplay()
	return dt.content
}

func (dt DirectoriesTab) UpdateDisplay() {
	// Empty the content first
	dt.content.Objects = nil

	// Empty state handling for directories
	emptyDirLabel := widget.NewLabel("No directories added yet.")
	dirListWrapper := container.NewVBox(emptyDirLabel)

	dirAddButton := widget.NewButton("Add Directory", func() {
		dt.SelectDirectory()
	})

	dirList := container.NewVBox()

	for _, dir := range dt.state.Directories {
		dt.logWriter.Write("Found directory " + dir.Path)
		label := widget.NewLabel(dir.Path)
		flowCountLabel := widget.NewLabel(fmt.Sprintf("(Flows: %d)", len(dir.Flows)))
		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

		removeBtn.OnTapped = func(d appstate.FlowDirectory) func() {
			return func() {
				dt.state.RemoveDirectory(d)                // Remove the directory
				dt.state.SaveToPreferences(dt.preferences) // Save updated preferences
				dt.UpdateDisplay()                         // Call updateDisplay after removal
			}
		}(dir)

		removeBtn.Disable()
		if len(dir.Flows) == 0 {
			removeBtn.Enable()
		}

		dirList.Add(container.NewHBox(label, flowCountLabel, removeBtn))
	}

	// Check if there are directories and adjust UI
	if len(dt.state.Directories) > 0 {
		dirListWrapper = container.NewVBox(dirAddButton, dirList)
	} else {
		dirListWrapper = container.NewVBox(dirAddButton, emptyDirLabel)
	}
	dt.content = dirListWrapper
	dt.content.Refresh()
}

func (dt DirectoriesTab) AddDirectory(dir string) {
	dt.state.AddDirectory(dir)
	dt.state.SaveToPreferences(dt.preferences)
	dt.UpdateDisplay()
}

// Function to open the file dialog for directory selection
func (dt DirectoriesTab) SelectDirectory() {
	var selectedDirectory string
	dirDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil || folder == nil {
			fmt.Println("No directory selected.")
			return
		}
		selectedDirectory = folder.String()
		// Here, the file URI will point to the directory selected
		dt.state.AddDirectory(selectedDirectory)
		dt.logWriter.Write(fmt.Sprintf("Directory selected: %s", selectedDirectory))
		dt.state.SaveToPreferences(dt.preferences)
	}, dt.window)

	dirDialog.Show()
}
