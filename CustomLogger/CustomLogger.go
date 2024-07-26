package CustomLogger

import (
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

type CustomLogger struct {
	logEntries []logEntry
	lock       sync.RWMutex
	maxLevel   int
}

func NewCustomLogger() CustomLogger {
	return CustomLogger{
		logEntries: make([]logEntry, 0),
		lock:       sync.RWMutex{},
		maxLevel:   INFO,
	}
}

func (cl *CustomLogger) LogError(message string, category string) {
	cl.logEntry(newLogEntry(message, ERROR, category))
}

func (cl *CustomLogger) LogWarning(message string, category string) {
	cl.logEntry(newLogEntry(message, WARNING, category))
}

func (cl *CustomLogger) LogInfo(message string, category string) {
	cl.logEntry(newLogEntry(message, INFO, category))
}

func (cl *CustomLogger) LogDebug(message string, category string) {
	cl.logEntry(newLogEntry(message, DEBUG, category))
}

func (cl *CustomLogger) logEntry(log logEntry) {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	cl.logEntries = append(cl.logEntries, log)
}

func (cl *CustomLogger) SetMinLevel(level int) {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	cl.maxLevel = level
}

func (cl *CustomLogger) IncreaseLogLevel() {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	if cl.maxLevel < DEBUG {
		cl.maxLevel++
	}
}

func (cl *CustomLogger) DecreaseLogLevel() {
	cl.lock.Lock()
	defer cl.lock.Unlock()
	if cl.maxLevel > ERROR {
		cl.maxLevel--
	}
}

func (cl *CustomLogger) GetCommandString() string {
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	keys := []string{}
	cmands := []string{}
	if cl.maxLevel > ERROR {
		keys = append(keys, "-")
		cmands = append(cmands, "Decrease")
	}
	if cl.maxLevel < DEBUG {
		keys = append(keys, "+")
		cmands = append(cmands, "Increase")
	}
	return strings.Join(keys, "/") + ": " + strings.Join(cmands, "/") + " log level [" + levelToString(cl.maxLevel) + "]"
}

func (cl *CustomLogger) GetLastNEntries(n int) []string {
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	entries := make([]string, 0)
	totalLen := len(cl.logEntries)
	insertedEntries := 0
	for i := totalLen - 1; i >= 0 && insertedEntries < n; i-- {
		entry := cl.logEntries[i]
		if entry.level <= cl.maxLevel {
			entries = append(entries, entry.String())
			insertedEntries++
		}
	}
	slices.Reverse(entries)
	return entries
}

func (cl *CustomLogger) RenderLogsWithLipGloss(width, n int) string {
	l := list.New(cl.GetLastNEntries(n)).Enumerator(func(items list.Items, index int) string {
		return ""
	})
	leftPadding := 2
	logs := lipgloss.NewStyle().Width(width - leftPadding).Border(lipgloss.RoundedBorder()).AlignVertical(lipgloss.Center).BorderForeground(lipgloss.Color("99")).PaddingLeft(leftPadding).Render(l.String())
	Title := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Width(width).Align(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, true).BorderForeground(lipgloss.Color("208")).
		Render("Latest Logs:")
	logs = lipgloss.JoinVertical(lipgloss.Left, Title, logs)
	return logs
}
