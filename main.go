package main

import (
	// Standard imports

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	// Local imports
	"fileflow/appstate"
	"fileflow/src/logutils"
	"fileflow/ui"

	// Third party imports
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/emersion/go-autostart"
)

const (
	appName = "FileFlow"
)

var logWriter *logutils.LogWriter

func main() {

	thisApp := app.New()
	appPreferences := thisApp.Preferences()

	thisApp.Settings().SetTheme(theme.DarkTheme())
	thisWindow := thisApp.NewWindow(fmt.Sprintf("%s Settings", appName))
	thisWindow.Resize(fyne.NewSize(600, 400))

	if a, ok := thisApp.(desktop.App); ok {
		fyneDesktopMenu := fyne.NewMenu(appName,
			fyne.NewMenuItem("Settings", func() {
				thisWindow.Show()
			}))
		a.SetSystemTrayMenu(fyneDesktopMenu)
	}

	logWriter, err := logutils.NewLogWriter(appName)
	if err != nil {
		logWriter.Fatal("Error with setting up logging interface!", err)
	}
	state := &appstate.AppState{}

	// Flows tab components

	flowAddButton := widget.NewButton("Add Flow", func() {
		if len(state.Directories) > 0 {
			showAddFlowDialog(thisWindow, state, &state.Directories[0]) // Assume the first directory is selected
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

	// Settings Tab
	startupToggle := widget.NewCheck("Start FileFlow on system startup", func(checked bool) {
		executable, _ := os.Executable()
		app := &autostart.App{
			Name:        "FileFlow",
			DisplayName: "FileFlow",
			Exec:        []string{executable},
		}
		if checked {
			thisApp.Preferences().SetBool("startup", true)
			if err := app.Enable(); err != nil {
				log.Fatal(err)
			}
			logWriter.Write("Enabled FileFlow on system startup")
		} else {
			thisApp.Preferences().SetBool("startup", false)
			if err := app.Disable(); err != nil {

				log.Fatal(err)
			}
			logWriter.Write("Disabled FileFlow on system startup")
		}
		// Logic to set startup preference
	})
	startupToggle.SetChecked(thisApp.Preferences().Bool("startup"))
	quitButton := widget.NewButton("Quit", func() {
		logWriter.Write("Quitting from Settings menu")
		thisApp.Quit()
	})
	settingsContent := container.NewVBox(startupToggle, layout.NewSpacer(), quitButton)

	directoriesTab := ui.NewDirectoriesTab(appPreferences, &appstate.AppState{}, thisWindow, logWriter)

	// Tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Introduction", createIntroductionTab()),
		container.NewTabItem("Directories", directoriesTab.GetContent()),
		container.NewTabItem("Flows", flowsListWrapper),
		container.NewTabItem("Logs", createLogViewerTab(logWriter.LogFilePath)),
		container.NewTabItem("Settings", settingsContent),
	)
	tabs.OnSelected = func(ti *container.TabItem) {
		logWriter.Write(fmt.Sprintf("Tab %s selected", ti.Text))
		if ti.Text == "Logs" {
			scrollContent := ti.Content.(*container.Scroll)
			scrollContent.ScrollToBottom()
		}
	}

	thisWindow.SetContent(tabs)

	thisWindow.SetCloseIntercept(func() {
		thisWindow.Hide()
	})
	thisWindow.ShowAndRun()

}

func createLogViewerTab(logFilePath string) *container.Scroll {

	logContent := widget.NewLabel("")
	logContent.Wrapping = fyne.TextWrapWord
	logContent.TextStyle = fyne.TextStyle{Monospace: true}
	scrollContainer := container.NewVScroll(logContent)

	scrollContainer.ScrollToBottom()

	refreshLogContent := func() {
		data, err := ioutil.ReadFile(logFilePath)
		if err != nil {
			logContent.SetText(fmt.Sprintf("Error reading log file %s: %v", logFilePath, err))
			return
		}
		logContent.SetText(string(data))

	}
	go func() {
		for range time.Tick(2 * time.Second) {
			refreshLogContent()
		}
	}()

	return scrollContainer
}

func createIntroductionTab() *fyne.Container {
	// Instruction text
	instructions := `Welcome to FileFlow!

FileFlow allows you to monitor specific directories for files that match user-defined patterns, called "Flows". 
Each Flow represents a rule, with a regular expression (regex) to identify files and a target directory 
to move those files into.

How to Use:
1. Adding Directories:
   - Use the "Add Directory" button to add directories for monitoring.
   - Directories must exist and will be validated before being added.

2. Defining Flows:
   - For each directory, define one or more Flows by specifying a regular expression (regex) and a target directory.
   - Regex patterns help identify which files to move based on their names.
   - Each Flow must have a valid regex pattern and a destination directory.

3. Managing Entries:
   - Edit or remove any Flow or directory as needed. 
   - A directory cannot be removed if it still contains active Flows.
   - Confirm all removals to prevent accidental deletions.

4. Viewing Logs:
   - Use the "Log Viewer" tab to see real-time log entries. This will show all actions taken by the application.

Start by adding a directory and defining your first Flow. Enjoy automating your file organization with FileFlow!
`

	// Create the label widget with instructions
	introLabel := widget.NewLabel(instructions)

	// Return the container to be used as the introduction tab
	return container.NewVBox(introLabel)
}

func isValidDirectory(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func showAddFlowDialog(win fyne.Window, state *appstate.AppState, dir *appstate.FlowDirectory) {
	flowPatternEntry := widget.NewEntry()
	flowPatternEntry.SetPlaceHolder("Enter regex pattern...")
	saveButton := widget.NewButton("Save", func() {
		pattern := flowPatternEntry.Text
		if pattern == "" || !isValidRegex(pattern) {
			return
		}
		//dir.Flows = append(dir.Flows, appstate.Flow{Pattern: pattern, Destination: dir})
	})
	win.SetContent(container.NewVBox(flowPatternEntry, saveButton))
}

func isValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}
