package appstate

import (
	"encoding/json"

	"fyne.io/fyne/v2"
)

type Flow struct {
	Pattern     string
	Destination DestinationDirectory
}

type Directory struct {
	Path string
}

type FlowDirectory struct {
	Directory
	Flows []Flow
}

type DestinationDirectory struct {
	Directory
}

type AppState struct {
	Directories []FlowDirectory
}

func NewAppState() *AppState {
	return &AppState{Directories: []FlowDirectory{}}
}

func (state *AppState) AddDirectory(path string) {
	state.Directories = append(state.Directories, FlowDirectory{Directory: Directory{Path: path}, Flows: []Flow{}})
}

func (state *AppState) AddFlowToDirectory(flowDir FlowDirectory, pattern string, destinationDir DestinationDirectory) {
	for i, dir := range state.Directories {
		if dir.Path == flowDir.Path {
			state.Directories[i].Flows = append(state.Directories[i].Flows, Flow{Pattern: pattern, Destination: destinationDir})
			break
		}
	}
}

func (state *AppState) RemoveDirectory(flowDir FlowDirectory) {
	if len(flowDir.Flows) == 0 {
		for i, dir := range state.Directories {
			if dir.Path == flowDir.Path {
				state.Directories = append(state.Directories[:i], state.Directories[i+1:]...)
				break
			}
		}
	}
}

func (state *AppState) SaveToPreferences(preferences fyne.Preferences) {
	data, err := json.Marshal(state.Directories)
	if err == nil {
		preferences.SetString("directories", string(data))
	}
}

func (state *AppState) LoadFromPreferences(preferences fyne.Preferences) {
	data := preferences.String("directories")
	if data != "" {
		json.Unmarshal([]byte(data), &state.Directories)
	}
}
