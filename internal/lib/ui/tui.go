package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/go-template/internal/lib/models"
)

const (
	leftMargin       = 4
	maxProgressWidth = 60
	pollPerMilli     = 60
)

var (
	borderColor = "#818C95"
	brightColor = "#A8FF00"
	baseStyle   = lipgloss.NewStyle().MarginLeft(leftMargin)

	headerStyle = baseStyle.Bold(true).Foreground(lipgloss.Color("#34C8FF"))
	noStyle     = baseStyle.Bold(true).Foreground(lipgloss.Color("#FFA31F"))
	brightStyle = baseStyle.Bold(true).Foreground(lipgloss.Color(brightColor))
	helpStyle   = baseStyle.Foreground(lipgloss.Color("#626262"))

	resultsHeaderStyle = brightStyle.Width(40).
				AlignHorizontal(lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(borderColor))

	resultsLabelStyle = baseStyle.
				AlignHorizontal(lipgloss.Right).
				Width(23).
				PaddingRight(1).
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(borderColor)).
				Foreground(lipgloss.Color("#DBEFFF"))

	resultsTImagesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34C8FF"))
	resultsValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFD2"))
	resultsDupeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD200"))
	resultsCacheStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E3BAFF"))
	resultsNewStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(brightColor))
	resultsTTimeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(brightColor))
	timeNotationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
)

type (
	MsgHashProgress    *models.ProcessStatus
	MsgUpdateProgress  *models.ProcessStatus
	MsgHashCompleted   bool
	MsgUpdateCompleted bool
	MsgDone            bool
)

type ResultDisplayItem struct {
	label      string
	value      string
	valueStyle lipgloss.Style
}

type TuiModel struct {
	selection             bool
	isSelected            bool
	isDone                bool
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
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", brightColor)),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", brightColor)),
		progressStatus:    &models.ProcessStatus{},
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
		pSize := maxProgressWidth - leftMargin
		if msg.Width < maxProgressWidth {
			pSize = msg.Width - leftMargin
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
	if !m.isSelected {
		return m.viewSelection()
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

func (m TuiModel) viewSelection() string {
	s := "\n" + headerStyle.Render("Would you like to process this directory?") + "\n\n"
	if !m.selection {
		s += noStyle.Render("> No") + "\n"
		s += baseStyle.Render("  Yes")
	} else {
		s += baseStyle.Render("  No") + "\n"
		s += brightStyle.Render("> Yes")
	}
	s += "\n\n" + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"
	return s
}

func (m TuiModel) viewProgress() string {
	margin := strings.Repeat(" ", leftMargin)
	s := ""

	if m.hashProgressPercent == 0 && m.updateProgressPercent == 0 {
		s = "\n" + brightStyle.Render("Getting Ready...") + "\n"
	}

	if m.hashProgressPercent > 0 {
		s = "\n" + brightStyle.Render("Hashing...") + "\n"
		s += "\n" + margin + m.hashProgressBar.ViewAs(m.hashProgressPercent) + "\n"
	}

	if m.updateProgressPercent > 0 {
		s += "\n" + brightStyle.Render("Updating...") + "\n"
		s += "\n" + margin + m.updateProgressBar.ViewAs(m.updateProgressPercent) + "\n"
	}

	s += "\n\n" + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"

	return s
}

func (m TuiModel) viewResults() string {
	s := fmt.Sprintf("\n%s\n\n", resultsHeaderStyle.Render("Hashimg Results"))

	items := []ResultDisplayItem{
		{"Total Images", strconv.Itoa(int(m.progressStatus.TotalImages)), resultsTImagesStyle},
		{"Dupes", strconv.Itoa(int(m.progressStatus.DupeImages)), resultsDupeStyle},
		{"Cached", strconv.Itoa(int(m.progressStatus.CachedImages)), resultsCacheStyle},
		{"New", strconv.Itoa(int(m.progressStatus.NewImages)), resultsNewStyle},
		{"", "", resultsValueStyle},
		{"Hash Speed", formatDuration(m.progressStatus.HashingTook), resultsValueStyle},
		{"Filter Speed", formatDuration(m.progressStatus.FilterTook), resultsValueStyle},
		{
			"Update Speed",
			formatDuration(m.progressStatus.UpdatingTook),
			resultsValueStyle,
		},
		{"", "", resultsValueStyle},
		{"Total Time", formatDuration(m.progressStatus.TotalTime), resultsTTimeStyle},
	}

	for _, item := range items {
		isInstant := strings.Contains(item.value, "0") &&
			strings.Contains(item.value, "ns") &&
			!strings.Contains(item.value, ".")

		// Ignore instantaneous speed results
		if isInstant {
			continue
		}
		s += fmt.Sprintf(
			"%s %s\n",
			resultsLabelStyle.Render(item.label),
			item.valueStyle.Render(item.value),
		)
	}

	return s
}

func formatDuration(d time.Duration) string {
	style := timeNotationStyle
	switch {
	case d >= time.Second:
		return fmt.Sprintf("%.2f"+style.Render("s"), d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%.2f"+style.Render("ms"), float64(d)/float64(time.Millisecond))
	case d >= time.Microsecond:
		return fmt.Sprintf("%.2f"+style.Render("µs"), float64(d)/float64(time.Microsecond))
	default:
		return fmt.Sprintf("%d"+style.Render("ns"), d.Nanoseconds())
	}
}
