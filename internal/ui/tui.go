package ui

import (
	"fmt"
	"reflect"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/hashimg/internal/models"
)

const (
	pollingRateMilli = 60
)

const (
	StateWelcome State = iota
	StateConsentSelection
	StateHDDSelection
	StateDoWork
	StateProgressing
	StateResults
	StateError
	StateDone
	StateAbort
)

const (
	ProgressHash ProgressState = iota
	ProgressHashComplete
	ProgressUpdate
	ProgressUpdateComplete
	ProgressHashErr
	ProgressUpdateErr
	ProgressDone
)

type (
	State         int
	ProgressState int
	MsgErr        struct {
		name string
		err  error
	}
)

type ResultDisplayItem struct {
	label      string
	value      string
	valueStyle lipgloss.Style
}

type WorkFunc = func(ps *models.ProcessStatus, useAvgBufferSize bool) error

type TuiModel struct {
	count                 int
	state                 State
	hasConsent            bool
	hddIndex              int
	hddList               []string
	isHDD                 bool
	workFunc              WorkFunc
	workErr               MsgErr
	hashProgressBar       progress.Model
	updateProgressBar     progress.Model
	hashProgressPercent   float64
	updateProgressPercent float64
	progressStatus        *models.ProcessStatus
}

func NewTUI(workFunc WorkFunc) TuiModel {
	return TuiModel{
		state:             StateConsentSelection,
		workFunc:          workFunc,
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", brightColor)),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", brightColor)),
		progressStatus:    &models.ProcessStatus{},
		workErr:           MsgErr{},
		hddList: []string{
			"HDD - Hard Disk Drive (Noisy)",
			"SSD - Solid State Drive (Flash)",
		},
	}
}

func (m TuiModel) Init() tea.Cmd {
	return nil
}

func (m TuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			m.state = StateAbort
		}

	case tea.WindowSizeMsg:
		pSize := maxProgressWidth - leftMargin
		if msg.Width < maxProgressWidth {
			pSize = msg.Width - leftMargin
		}
		m.hashProgressBar.Width = pSize
		m.updateProgressBar.Width = pSize
	}

	switch m.state {

	case StateAbort:
		return m, tea.Quit

	case StateConsentSelection:
		return m.updateConsentSelection(msg)

	case StateHDDSelection:
		return m.updateHDDSelection(msg)

	case StateDoWork:
		m.state = StateProgressing
		go m.workFunc(m.progressStatus, m.isHDD)
		return m, m.pollProgressStatus()

	case StateProgressing:
		return m.updateProgress(msg)

	default:
		panic(fmt.Sprintf("unknown state: %d", m.state))

	}
}

func (m TuiModel) View() string {
	switch m.state {

	case StateAbort:
		return m.viewAbort()

	case StateError:
		return m.viewErr(m.workErr)

	case StateConsentSelection:
		return m.viewConsentSelection()

	case StateHDDSelection:
		return m.viewHardDriveSelection()

	case StateDoWork, StateProgressing:
		return m.viewProgress()

	case StateResults:
		return m.viewResults()

	}

	panic("missing view")
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
			if !m.hasConsent {
				m.state = StateAbort
				return m.Update(msg)
			}
			m.state = StateHDDSelection
			return m, nil
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
			m.state = StateDoWork
			m.isHDD = m.hddIndex == 0
			return m.Update(msg)
		}
	}
	return m, nil
}

func (m TuiModel) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case ProgressState:
		switch msg {

		case ProgressHash:
			progressBy := 100 / float64(m.progressStatus.MaxHashProgress)
			m.hashProgressPercent = progressBy / 100 * float64(m.progressStatus.HashProgress)
			return m, m.pollProgressStatus()

		case ProgressHashComplete:
			m.hashProgressPercent = 1
			return m, m.pollProgressStatus()

		case ProgressUpdate:
			progressBy := 100 / float64(m.progressStatus.MaxUpdateProgress)
			m.updateProgressPercent = progressBy / 100 * float64(m.progressStatus.UpdateProgress)
			return m, m.pollProgressStatus()

		case ProgressUpdateComplete:
			m.updateProgressPercent = 1
			return m, m.pollProgressStatus()

		case ProgressHashErr:
			m.state = StateError
			m.workErr.name = "Hashing"
			m.workErr.err = m.progressStatus.HashErr
			return m, tea.Quit

		case ProgressUpdateErr:
			m.state = StateError
			m.workErr.name = "Updating"
			m.workErr.err = m.progressStatus.UpdateErr
			return m, tea.Quit

		case ProgressDone:
			m.state = StateResults
			return m, tea.Quit

		default:
			panic(fmt.Sprintf("missing progress state: %d", msg))

		}

	default:
		panic(fmt.Sprintf("invalid type for progress update: %s", reflect.TypeOf(msg).Name()))

	}
}

func (m TuiModel) pollProgressStatus() tea.Cmd {
	return tea.Tick(time.Millisecond*pollingRateMilli, func(t time.Time) tea.Msg {
		if m.progressStatus.HashErr != nil {
			return ProgressHashErr
		}

		if m.progressStatus.UpdateErr != nil {
			return ProgressUpdateErr
		}

		if m.progressStatus.HashProgress != m.progressStatus.MaxHashProgress {
			return ProgressHash
		}

		if m.hashProgressPercent != 1 {
			return ProgressHashComplete
		}

		if m.progressStatus.UpdateProgress != m.progressStatus.MaxUpdateProgress {
			return ProgressUpdate
		}

		if m.updateProgressPercent != 1 {
			return ProgressUpdateComplete
		}

		if m.updateProgressPercent == 1 {
			return ProgressDone
		}

		panic("tried to send empty progress state")
	})
}
