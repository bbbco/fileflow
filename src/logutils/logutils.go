package logutils

import (
	"fyne.io/fyne/v2/widget"
)

type LogWriter struct {
	LogContent *widget.Entry
}

func NewLogWriter(logContent *widget.Entry) *LogWriter {
	return &LogWriter{
		LogContent: logContent,
	}
}

func (lw *LogWriter) Write(p string) (n int, err error) {
	// Append the new log message to the current content
	lw.LogContent.SetText(lw.LogContent.Text + string("\n") + p)
	// Scroll to the bottom to show the latest log
	lw.LogContent.CursorRow = -1
	return len(p), nil
}
