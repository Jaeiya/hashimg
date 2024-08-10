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
	CautionForeColor = "#FFF10E"
	CautionBackColor = "#131313"

	CautionStyle = baseStyle.
			Width(40).
			AlignHorizontal(lipgloss.Center).
			MarginTop(2).
			MarginLeft(4).
			Padding(1, 2).
			Background(lipgloss.Color(CautionBackColor)).
			Foreground(lipgloss.Color(CautionForeColor))

	borderColor = "#818C95"
	brightColor = "#A8FF00"
	baseStyle   = lipgloss.NewStyle().MarginLeft(leftMargin)

	headerStyle = baseStyle.Bold(true).Foreground(lipgloss.Color("#34C8FF"))
	noStyle     = baseStyle.Bold(true).Foreground(lipgloss.Color("#FFA31F"))
	brightStyle = baseStyle.Bold(true).Foreground(lipgloss.Color(brightColor))
	footerStyle = baseStyle.Foreground(lipgloss.Color("#626262"))

	resultsHeaderStyle = brightStyle.
				MarginLeft(3).
				Width(30).
				Padding(1, 0).
				AlignHorizontal(lipgloss.Center).
				Background(lipgloss.Color("#003284"))

	errorHeaderStyle = baseStyle.
				Width(40).
				Padding(1, 2).
				AlignHorizontal(lipgloss.Center).
				Foreground(lipgloss.Color("#FF71CB")).
				Background(lipgloss.Color(CautionBackColor))

	resultsLabelStyle = baseStyle.
				AlignHorizontal(lipgloss.Right).
				Width(16).
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
	MsgEmpty           bool
	MsgErr             struct {
		name string
		err  error
	}
)

type ResultDisplayItem struct {
	label      string
	value      string
	valueStyle lipgloss.Style
}

type TuiModel struct {
	hasConsent            bool
	hasSelectedConsent    bool
	hddIndex              int
	hddList               []string
	hasSelectedHDD        bool
	isDone                bool
	workFunc              func(ps *models.ProcessStatus, useAvgBufferSize bool)
	isWorking             bool
	workErr               MsgErr
	hashProgressBar       progress.Model
	updateProgressBar     progress.Model
	hashProgressPercent   float64
	updateProgressPercent float64
	progressStatus        *models.ProcessStatus
	footerText            string
}

func NewTUI(
	appVersion string,
	workFunc func(ps *models.ProcessStatus, useAvgBufferSize bool),
) TuiModel {
	return TuiModel{
		workFunc:          workFunc,
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", brightColor)),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", brightColor)),
		progressStatus:    &models.ProcessStatus{},
		workErr:           MsgErr{},
		hddList: []string{
			"HDD - Hard Disk Drive (Noisy)",
			"SSD - Solid State Drive (Flash)",
		},
		footerText: footerStyle.Render(
			"Hashimg " + appVersion + " - Press Esc or Ctrl+C to quit",
		),
	}
}

func (m TuiModel) Init() tea.Cmd {
	return nil
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		pSize := maxProgressWidth - leftMargin
		if msg.Width < maxProgressWidth {
			pSize = msg.Width - leftMargin
		}
		m.hashProgressBar.Width = pSize
		m.updateProgressBar.Width = pSize
	}

	if !m.hasSelectedConsent {
		return m.updateConsentSelection(msg)
	}

	if !m.hasSelectedHDD {
		return m.updateHDDSelection(msg)
	}

	if !m.isWorking {
		isHDD := false
		if m.hddList[m.hddIndex] == "HDD" {
			isHDD = true
		}
		m.isWorking = true
		go m.workFunc(m.progressStatus, isHDD)
		return m, m.pollUpdates()
	}

	return m.updateProgress(msg)
}

func (m TuiModel) View() string {
	if m.hasSelectedConsent && !m.hasConsent {
		return m.viewCancel()
	}

	if m.workErr.err != nil {
		return m.viewErr(m.workErr)
	}

	if !m.hasSelectedConsent {
		return m.viewSelection()
	}

	if m.hasConsent && !m.hasSelectedHDD {
		return m.viewHardDriveSelection()
	}

	if m.hasConsent && !m.isDone {
		return m.viewProgress()
	}

	if m.isDone {
		return m.viewResults()
	}

	return "Oops, something went wrong"
}

func (m TuiModel) updateConsentSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.hasConsent = false

		case "down", "j":
			m.hasConsent = true

		case "enter":
			m.hasSelectedConsent = true
			if !m.hasConsent {
				return m, tea.Quit
			}
			return m, func() tea.Msg {
				return MsgEmpty(true)
			}
		}
	}
	return m, nil
}

func (m TuiModel) updateHDDSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.hddIndex--
			if m.hddIndex < 0 {
				m.hddIndex = 0
			}

		case "down", "j":
			m.hddIndex++
			listLen := len(m.hddList)
			if m.hddIndex >= listLen {
				m.hddIndex = listLen - 1
			}

		case "enter":
			m.hasSelectedHDD = true
			return m, func() tea.Msg {
				return MsgEmpty(true)
			}
		}
	}
	return m, nil
}

func (m TuiModel) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

	case MsgErr:
		m.workErr = msg
		return m, tea.Quit

	case MsgDone:
		m.isDone = true
		return m, tea.Quit

	default:
		return m, nil
	}
}

func (m TuiModel) pollUpdates() tea.Cmd {
	return tea.Tick(time.Millisecond*pollPerMilli, func(t time.Time) tea.Msg {
		if m.progressStatus.HashErr != nil {
			return MsgErr{"Hashing", m.progressStatus.HashErr}
		}

		if m.progressStatus.UpdateErr != nil {
			return MsgErr{"Updating", m.progressStatus.UpdateErr}
		}

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
	if !m.hasConsent {
		s += noStyle.Render("> No") + "\n"
		s += baseStyle.Render("  Yes")
	} else {
		s += baseStyle.Render("  No") + "\n"
		s += brightStyle.Render("> Yes")
	}
	s += "\n\n" + m.footerText + "\n"
	return s
}

func (m TuiModel) viewHardDriveSelection() string {
	s := "\n" + headerStyle.Render("What kind of hard drive are your files on?") + "\n\n"
	for i, hd := range m.hddList {
		if m.hddIndex == i {
			s += brightStyle.Render("> "+hd) + "\n"
		} else {
			s += baseStyle.Render("  "+hd) + "\n"
		}
	}
	s += "\n" + m.footerText + "\n"
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

	s += "\n\n" + m.footerText + "\n"

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
		{"Analyze Speed", formatDuration(m.progressStatus.AnalyzeTook), resultsValueStyle},
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

func (m TuiModel) viewErr(e MsgErr) string {
	s := "\n" + errorHeaderStyle.Render("Error Occurred During "+e.name) + "\n\n"

	s += baseStyle.Width(60).
		Foreground(lipgloss.Color("#E3F7FF")).
		Render(fmt.Sprintf("%s", e.err)) +
		"\n"
	return s
}

func (m TuiModel) viewCancel() string {
	s := CautionStyle.Render("Aborted by User") + "\n"
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
