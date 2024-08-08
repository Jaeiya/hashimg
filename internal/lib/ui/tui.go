package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/go-template/internal/lib/models"
)

const (
	padding          = 4
	maxProgressWidth = 60
	pollPerMilli     = 60
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#34C8FF"))
	noStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA31F"))
	yesStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A8FF00"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	instantStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#FFC000"))
)

type (
	MsgHashProgress    *models.ProcessStatus
	MsgUpdateProgress  *models.ProcessStatus
	MsgHashCompleted   bool
	MsgUpdateCompleted bool
	MsgDone            bool
)

type TuiModel struct {
	selection             bool
	isSelected            bool
	isDone                bool
	padding               string
	workFunc              func(ps *models.ProcessStatus)
	hashProgressBar       progress.Model
	updateProgressBar     progress.Model
	hashProgressPercent   float64
	updateProgressPercent float64
	progressStatus        *models.ProcessStatus
}

func NewTUI(workFunc func(ps *models.ProcessStatus)) TuiModel {
	return TuiModel{
		workFunc:          workFunc,
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", "#A8FF00")),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", "#A8FF00")),
		progressStatus:    &models.ProcessStatus{},
		padding:           strings.Repeat(" ", padding),
	}
}

func (m TuiModel) Init() tea.Cmd {
	return nil
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeys(msg)

	case tea.WindowSizeMsg:
		pSize := maxProgressWidth - padding
		if msg.Width < maxProgressWidth {
			pSize = msg.Width - padding
		}
		m.hashProgressBar.Width = pSize
		m.updateProgressBar.Width = pSize

	case MsgHashProgress:
		progressBy := 100 / float64(msg.MaxHashProgress)
		m.hashProgressPercent = progressBy / 100 * float64(msg.HashProgress)
		return m, m.pollUpdates()

	case MsgUpdateProgress:
		progressBy := 100 / float64(msg.MaxUpdateProgress)
		m.updateProgressPercent = progressBy / 100 * float64(msg.UpdateProgress)
		return m, m.pollUpdates()

	case MsgUpdateCompleted:
		m.updateProgressPercent = 1
		return m, m.pollUpdates()

	case MsgHashCompleted:
		m.hashProgressPercent = 1
		return m, m.pollUpdates()

	case MsgDone:
		m.isDone = true
		return m, tea.Quit

	default:
		return m, nil
	}

	return m, nil
}

func (m TuiModel) View() string {
	pad := strings.Repeat(" ", padding)
	s := "\n" + headerStyle.Render(
		fmt.Sprintf("%sWould you like to process this directory?", pad),
	) + "\n\n"

	if !m.isSelected {
		return m.viewSelection(s, pad)
	}

	if m.selection && !m.isDone {
		return m.viewProgress()
	}

	if m.isDone {
		return m.viewResults()
	}

	return "\nCancelled"
}

func (m TuiModel) handleKeys(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.selection = false

	case "down", "j":
		m.selection = true

	case "enter":
		m.isSelected = true
		if m.selection {
			go m.workFunc(m.progressStatus)
			return m, m.pollUpdates()
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m TuiModel) pollUpdates() tea.Cmd {
	return tea.Tick(time.Millisecond*pollPerMilli, func(t time.Time) tea.Msg {
		if m.progressStatus.HashProgress != m.progressStatus.MaxHashProgress {
			return MsgHashProgress(m.progressStatus)
		}

		if m.hashProgressPercent != 1 {
			return MsgHashCompleted(true)
		}

		if m.progressStatus.UpdateProgress != m.progressStatus.MaxUpdateProgress {
			return MsgUpdateProgress(m.progressStatus)
		}

		if m.updateProgressPercent != 1 {
			return MsgUpdateCompleted(true)
		}

		if m.updateProgressPercent == 1 {
			return MsgDone(true)
		}

		return m.progressStatus
	})
}

func (m *TuiModel) viewSelection(s string, padding string) string {
	selNo, selYes := " ", " "
	if !m.selection {
		selNo = ">"
		s += noStyle.Render(fmt.Sprintf("%s%s %s", padding, selNo, "No"))
		s += fmt.Sprintf("\n%s  %s", padding, "Yes")
	} else {
		selYes = ">"
		s += fmt.Sprintf("%s  %s\n", padding, "No")
		s += yesStyle.Render(fmt.Sprintf("%s%s %s", padding, selYes, "Yes"))
	}
	s += "\n\n" + padding + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"
	return s
}

func (m *TuiModel) viewProgress() string {
	pad := m.padding
	s := ""

	if m.hashProgressPercent == 0 && m.updateProgressPercent == 0 {
		s = "\n" + pad + yesStyle.Render("Getting Ready...\n")
	}

	if m.hashProgressPercent > 0 {
		s = "\n" + pad + yesStyle.Render("Hashing...\n")
		s += "\n" + pad + m.hashProgressBar.ViewAs(m.hashProgressPercent) + "\n"
	}

	if m.updateProgressPercent > 0 {
		s += "\n" + pad + yesStyle.Render("Updating...\n")
		s += "\n" + pad + m.updateProgressBar.ViewAs(m.updateProgressPercent) + "\n"
	}

	s += "\n\n" + pad + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"

	return s
}

func (m *TuiModel) viewResults() string {
	pad := m.padding
	s := ""
	s = fmt.Sprintf("\n%s%s\n\n", pad, yesStyle.Render("Hashimg Results"))
	s += fmt.Sprintf("%s     Images: %d\n", pad, m.progressStatus.TotalImages)
	s += fmt.Sprintf("%s      Dupes: %d\n", pad, m.progressStatus.DupeImages)
	s += fmt.Sprintf("%s     Cached: %d\n", pad, m.progressStatus.CachedImages)
	s += fmt.Sprintf("%s     Hash Speed: %s\n", pad, m.progressStatus.HashingTook)
	s += fmt.Sprintf("%s     Update Speed: %s\n", pad, m.progressStatus.UpdatingTook)

	ft := time.Duration.String(m.progressStatus.FilterTook)
	filterStyle := ft
	if ft == "0s" {
		filterStyle = instantStyle.Render("Instant")
	}
	s += fmt.Sprintf("%s     Filter Speed: %s\n", pad, filterStyle)

	return s
}
