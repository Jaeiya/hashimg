package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/hashimg/internal/utils"
)

const (
	leftMargin       = 4
	maxProgressWidth = 60

	CautionForeColor = "#FFF10E"
	CautionBackColor = "#131313"

	borderColor = "#818C95"
	brightColor = "#A8FF00"
	darkColor   = "#626262"
	whiteColor  = "#DBEFFF"
	redColor    = "#FF71CB"

	hddSelectionText = "For performance reasons, selecting the kind of hard drive" +
		" your images are stored on, will allow the program to optimize itself.\n\n" +
		"In some cases, this can speed up the process from 10 seconds, to only taking" +
		" 6 seconds.\n\n" +
		"The more images you have, the more seconds you'll save. If you have less than" +
		" 100 images, you probably won't notice much difference between the two options."

	welcomeConsentText = "Welcome to Hashimg!\n\n" +
		"All images in the current working directory, will be compared for duplicates and" +
		" renamed to their truncated 32-character sha256 hash.\n\n" +
		"Renaming the images ensures that only new images will need to be fully processed."
)

var (
	footerText = footerStyle.Render(
		"Hashimg " + utils.AppVersion + " - Press Esc, Ctrl+C, or Q to quit",
	)

	CautionStyle = baseStyle.
			Width(40).
			AlignHorizontal(lipgloss.Center).
			MarginTop(2).
			MarginLeft(4).
			Padding(1, 2).
			Background(lipgloss.Color(CautionBackColor)).
			Foreground(lipgloss.Color(CautionForeColor))

	baseStyle = lipgloss.NewStyle().MarginLeft(leftMargin)

	headerStyle = baseStyle.Width(60).Bold(true).Foreground(lipgloss.Color("#34C8FF"))
	noStyle     = baseStyle.Bold(true).Foreground(lipgloss.Color(redColor))
	brightStyle = baseStyle.Bold(true).Foreground(lipgloss.Color(brightColor))
	footerStyle = baseStyle.Foreground(lipgloss.Color(darkColor))

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
				Foreground(lipgloss.Color(redColor)).
				Background(lipgloss.Color(CautionBackColor))

	resultsLabelStyle = baseStyle.
				AlignHorizontal(lipgloss.Right).
				Width(16).
				PaddingRight(1).
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(borderColor)).
				Foreground(lipgloss.Color(whiteColor))

	resultsTImagesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34C8FF"))
	resultsValueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFD2"))
	resultsDupeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD200"))
	resultsCacheStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E3BAFF"))
	resultsNewStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(brightColor))
	resultsTTimeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(brightColor))
	timeNotationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor))
)

func (m TuiModel) viewConsentSelection() string {
	s := "\n" + headerStyle.Render(welcomeConsentText) + "\n\n"
	s += baseStyle.Foreground(lipgloss.Color(whiteColor)).
		Render("Would you like to continue?") +
		"\n\n"
	if !m.hasConsent {
		s += noStyle.Render("> No") + "\n"
		s += baseStyle.Render("  Yes")
	} else {
		s += baseStyle.Render("  No") + "\n"
		s += brightStyle.Render("> Yes")
	}
	s += "\n\n" + footerText + "\n"
	return s
}

func (m TuiModel) viewHardDriveSelection() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	s := "\n" + headerStyle.Render(hddSelectionText) + "\n\n"

	driveStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color(brightColor)).
		Render(wd[:3])

	s += baseStyle.Foreground(lipgloss.Color(whiteColor)).
		Render("Which type of drive is your "+driveStr+" drive?") + "\n\n"

	for i, hd := range m.hddList {
		if m.hddIndex == i {
			s += brightStyle.Render("> "+hd) + "\n"
		} else {
			s += baseStyle.Render("  "+hd) + "\n"
		}
	}
	s += "\n" + footerText + "\n"
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

	s += fmt.Sprintf("\n\n%s %d\n", footerText, m.count)

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
		Foreground(lipgloss.Color(whiteColor)).
		Render(fmt.Sprintf("%s", e.err)) +
		"\n"
	return s
}

func (m TuiModel) viewAbort() string {
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
