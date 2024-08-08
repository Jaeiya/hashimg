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
	padding          = 4
	maxProgressWidth = 60
	pollPerMilli     = 60
)

var (
	borderColor = "#848994"
	brightColor = "#A8FF00"
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#34C8FF"))
	noStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA31F"))
	brightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(brightColor))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	resultsHeaderStyle = brightStyle.Width(40).
				AlignHorizontal(lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(borderColor))

	resultsLabelStyle = lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Right).
				Width(22).
				PaddingLeft(padding).
				PaddingRight(1).
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(borderColor)).
				Foreground(lipgloss.Color("#FFF"))

	resultsTImagesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34C8FF"))
	resultsValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFD2"))
	resultsDupeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD200"))
	resultsCacheStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E3BAFF"))
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
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", brightColor)),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", brightColor)),
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

func (m TuiModel) viewSelection(s string, padding string) string {
	selNo, selYes := " ", " "
	if !m.selection {
		selNo = ">"
		s += noStyle.Render(fmt.Sprintf("%s%s %s", padding, selNo, "No"))
		s += fmt.Sprintf("\n%s  %s", padding, "Yes")
	} else {
		selYes = ">"
		s += fmt.Sprintf("%s  %s\n", padding, "No")
		s += brightStyle.Render(fmt.Sprintf("%s%s %s", padding, selYes, "Yes"))
	}
	s += "\n\n" + padding + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"
	return s
}

func (m TuiModel) viewProgress() string {
	pad := m.padding
	s := ""

	if m.hashProgressPercent == 0 && m.updateProgressPercent == 0 {
		s = "\n" + pad + brightStyle.Render("Getting Ready...\n")
	}

	if m.hashProgressPercent > 0 {
		s = "\n" + pad + brightStyle.Render("Hashing...\n")
		s += "\n" + pad + m.hashProgressBar.ViewAs(m.hashProgressPercent) + "\n"
	}

	if m.updateProgressPercent > 0 {
		s += "\n" + pad + brightStyle.Render("Updating...\n")
		s += "\n" + pad + m.updateProgressBar.ViewAs(m.updateProgressPercent) + "\n"
	}

	s += "\n\n" + pad + helpStyle.Render("Hashimg 1.0 - Press q or ctrl+c to quit") + "\n"

	return s
}

type Stat struct {
	label      string
	value      string
	valueStyle lipgloss.Style
}

func (m TuiModel) viewResults() string {
	s := "\n"
	s += fmt.Sprintf("%s\n\n", resultsHeaderStyle.Render("Hashimg Results"))

	stats := []Stat{
		{"Total Images", strconv.Itoa(int(m.progressStatus.TotalImages)), resultsTImagesStyle},
		{"Dupes", strconv.Itoa(int(m.progressStatus.DupeImages)), resultsDupeStyle},
		{"Cached", strconv.Itoa(int(m.progressStatus.CachedImages)), resultsCacheStyle},
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

	for _, stat := range stats {
		isInstant := strings.Contains(stat.value, "0") &&
			strings.Contains(stat.value, "ns") &&
			!strings.Contains(stat.value, ".")

		// Ignore instantaneous speed results
		if isInstant {
			continue
		}
		s += fmt.Sprintf(
			"%s %s\n",
			resultsLabelStyle.Render(stat.label),
			stat.valueStyle.Render(stat.value),
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
