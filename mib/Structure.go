package mib

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/eivarin/LSNMPvS-DomoticSystem/CustomLogger"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
)

type Structure struct {
	Name         string
	StructureIID int
	Description  string
	Logger 	 	*CustomLogger.CustomLogger
	lock         *sync.RWMutex
}

func NewStructure(Name string, StructureIID int, Description string) Structure {
	return Structure{
		Name:         Name,
		StructureIID: StructureIID,
		Description:  Description,
		lock:         &sync.RWMutex{},
	}
}

func (s *Structure) Lock() {
	s.lock.Lock()
}

func (s *Structure) Unlock() {
	s.lock.Unlock()
}

func (s *Structure) RLock() {
	s.lock.RLock()
}

func (s *Structure) RUnlock() {
	s.lock.RUnlock()
}

func (s *Structure) renderStructureTableWithLipGloss(Titles []string, Values [][]string, width int) string {
	ansiClr := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	t := table.New().Border(lipgloss.RoundedBorder()).BorderStyle(ansiClr).Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
				case row == 0:
					return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
				case row%2 == 0:
					return lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
				default:
					return lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
			}
		}).Headers(Titles...).
		Rows(Values...)
	StructTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Width(width).Align(lipgloss.Center).
		Border(lipgloss.NormalBorder(), false, false, true).BorderForeground(lipgloss.Color("208")).
		Render("Structure: " + s.Name)
	return lipgloss.JoinVertical(lipgloss.Center, StructTitle, t.Render())
}

type StructureI interface {
	Get(objectIID, index int) (*types.CompleteCodableValue, packet.PacketErr)
	GetStructureName() string
	GetStructureIID() int
	GetDescription() string
	Set(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr
	Update(objectIID, index int, value types.CompleteCodableValue)
	PopulateObjectIDWithLength(objectIID int, length int)
	RenderTableWithLipGloss(width int) string
	Len() int
	Count(objectIID int) int
	GetDimensions() map[int]int
}
