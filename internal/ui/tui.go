package ui

import (
	"fmt"
	"reflect"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/hashimg/internal"
)

const (
	pollingRateMilli = 60
)

const (
	StateWelcome State = iota
	StateConsentSelection
	StateHDDSelection
	StateReviewConsentSelection
	StateUserReview
	StateDoAllWork
	StateDoHashWork
	StateDoUpdateWork
	StateDoHashReviewWork
	StateDoUpdateReviewWork
	StateHashProgressing
	StateUpdateProgressing
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

type TuiModel struct {
	state                 State
	hasConsent            bool
	wantsReview           bool
	keepDupes             bool
	hddIndex              int
	isHDD                 bool
	imgProcessor          *internal.ImageProcessor
	workErr               MsgErr
	hashProgressBar       progress.Model
	updateProgressBar     progress.Model
	hashProgressPercent   float64
	updateProgressPercent float64
}

func NewTUI(ip *internal.ImageProcessor) TuiModel {
	return TuiModel{
		state:             StateConsentSelection,
		hashProgressBar:   progress.New(progress.WithGradient("#34C8FF", brightColor)),
		updateProgressBar: progress.New(progress.WithGradient("#34C8FF", brightColor)),
		workErr:           MsgErr{},
		imgProcessor:      ip,
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

	case StateAbort, StateError:
		return m, tea.Quit

	case StateConsentSelection:
		return m.updateConsentSelection(msg)

	case StateHDDSelection:
		return m.updateHDDSelection(msg)

	case StateReviewConsentSelection:
		return m.updateReviewConsentSelection(msg)

	case StateDoHashWork:
		m.state = StateHashProgressing
		go m.imgProcessor.ProcessImages(m.isHDD)
		return m, m.pollProgressStatus()

	case StateDoUpdateWork:
		m.state = StateUpdateProgressing
		go m.imgProcessor.UpdateImages()
		return m, m.pollProgressStatus()

	case StateDoHashReviewWork:
		m.state = StateHashProgressing
		go m.imgProcessor.ProcessImagesForReview(m.isHDD)
		return m, m.pollProgressStatus()

	case StateDoUpdateReviewWork:
		m.state = StateUpdateProgressing
		err := m.imgProcessor.RestoreFromReview()
		if err != nil {
			m.state = StateError
			m.workErr.err = err
			return m.Update(msg)
		}
		go m.imgProcessor.UpdateImages()
		return m, m.pollProgressStatus()

	case StateUserReview:
		return m.updateUserReviewSelection(msg)

	case StateHashProgressing, StateUpdateProgressing:
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

	case StateReviewConsentSelection:
		return m.viewReviewConsentSelection()

	case StateDoAllWork,
		StateDoHashWork,
		StateDoUpdateWork,
		StateHashProgressing,
		StateUpdateProgressing:
		return m.viewProgress()

	case StateUserReview:
		return m.viewUserReviewSelection()

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
			// There are only two kinds of hard drives
			listLen := 2
			if m.hddIndex >= listLen {
				m.hddIndex = listLen - 1
			}

		case "enter":
			m.isHDD = m.hddIndex == 0
			m.state = StateReviewConsentSelection
			return m, nil
		}
	}
	return m, nil
}

func (m TuiModel) updateReviewConsentSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.wantsReview = false

		case "down", "j":
			m.wantsReview = true

		case "enter":
			if m.wantsReview {
				m.state = StateDoHashReviewWork
				return m.Update(msg)
			}
			m.state = StateDoHashWork
			return m.Update(msg)
		}
	}
	return m, nil
}

func (m TuiModel) updateUserReviewSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.keepDupes = false

		case "down", "j":
			m.keepDupes = true

		case "enter":
			if m.keepDupes {
				m.state = StateAbort
				return m.Update(msg)
			}
			m.state = StateDoUpdateReviewWork
			return m.Update(msg)
		}
	}
	return m, nil
}

func (m TuiModel) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	status := m.imgProcessor.Status
	switch msg.(type) {
	case ProgressState:
		switch msg {

		case ProgressHash:
			progressBy := 100 / float64(status.MaxHashProgress)
			m.hashProgressPercent = progressBy / 100 * float64(status.HashProgress)
			return m, m.pollProgressStatus()

		case ProgressHashComplete:
			m.hashProgressPercent = 1
			if m.wantsReview && m.imgProcessor.HasDupes {
				m.state = StateUserReview
				return m.Update(msg)
			}
			m.state = StateDoUpdateWork
			return m.Update(msg)

		case ProgressUpdate:
			progressBy := 100 / float64(status.MaxUpdateProgress)
			m.updateProgressPercent = progressBy / 100 * float64(status.UpdateProgress)
			return m, m.pollProgressStatus()

		case ProgressUpdateComplete:
			m.updateProgressPercent = 1
			m.state = StateResults
			return m, tea.Quit

		case ProgressHashErr:
			m.state = StateError
			m.workErr.name = "Hashing"
			m.workErr.err = status.HashErr
			return m, tea.Quit

		case ProgressUpdateErr:
			m.state = StateError
			m.workErr.name = "Updating"
			m.workErr.err = status.UpdateErr
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
	status := m.imgProcessor.Status
	return tea.Tick(time.Millisecond*pollingRateMilli, func(t time.Time) tea.Msg {
		if status.HashErr != nil {
			return ProgressHashErr
		}

		if status.UpdateErr != nil {
			return ProgressUpdateErr
		}

		if status.HashProgress != status.MaxHashProgress {
			return ProgressHash
		}

		if status.ProcessingComplete && m.state == StateHashProgressing {
			return ProgressHashComplete
		}

		if status.UpdateProgress != status.MaxUpdateProgress {
			return ProgressUpdate
		}

		if status.UpdatingComplete && m.state == StateUpdateProgressing {
			return ProgressUpdateComplete
		}

		panic("tried to send empty progress state")
	})
}
