package CustomLogger

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	ERROR = iota
	WARNING
	INFO
	DEBUG 
)

type logEntry struct {
	level    int
	category string
	message  string
	logTime  time.Time
}

func newLogEntry(message string, level int, category string) logEntry {
	return logEntry{
		level:    level,
		category: category,
		message:  message,
		logTime:  time.Now(),
	}
}

func (l *logEntry) String() string {
	catColor := levelToColor(l.level)
	lvlStr := levelToString(l.level)
	return l.logTime.Format("2006-01-02 15:04:05") + " {" + lvlStr + "} " + " [" + lipgloss.NewStyle().Foreground(lipgloss.Color(catColor)).Render(l.category) + "] " + l.message
}

func levelToString(level int) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	}
	return "UNKNOWN"
}

func levelToColor(level int) string {
	switch level {
	case DEBUG:
		return "240"
	case INFO:
		return "33"
	case WARNING:
		return "220"
	case ERROR:
		return "196"
	}
	return "0"
}