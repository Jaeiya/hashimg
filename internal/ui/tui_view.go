package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/hashimg/internal"
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

	welcomeConsentText = "Welcome to Hashimg!\n\n" +
		"All images in the current working directory, will be compared for duplicates and" +
		" renamed to their truncated 32-character sha256 hash.\n\n" +
		"Renaming the images ensures that only new images will need to be fully processed."

	hddSelectionText = "For performance reasons, selecting the kind of hard drive" +
		" your images are stored on, will allow the program to optimize itself.\n\n" +
		"In some cases, this can speed up the process from 10 seconds, to only taking" +
		" 6 seconds.\n\n" +
		"The more images you have, the more seconds you'll save. If you have less than" +
		" 100 images, you probably won't notice much difference between the two options."

	reviewSelectionText = "If you would like, you can review the duplicate images before" +
		" they are deleted. A separate folder will be opened for you to review them."

	userReviewSelectionText = "I've opened a folder containing the duplicate images." +
		" Images with the same hash name are considered identical. When you're done" +
		" reviewing the images, you can decide to keep or delete them below."
)

var (
	driveTextList = []string{
		"HDD - Hard Disk Drive (Noisy)",
		"SSD - Solid State Drive (Flash)",
	}

	footerText = fmt.Sprintf(
		"\n%s %s %s",
		footerStyle.Render("Hashimg"),
		getPrettyVersion(),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(darkColor)).
			Render("- Press Esc, Ctrl+C, or Q to quit"),
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
	return viewYesNo(
		"Would you like to continue?",
		welcomeConsentText,
		func() bool { return m.hasConsent },
	)
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

	for i, hd := range driveTextList {
		if m.hddIndex == i {
			s += brightStyle.Render("> "+hd) + "\n"
		} else {
			s += baseStyle.Render("  "+hd) + "\n"
		}
	}
	s += "\n" + footerText + "\n"
	return s
}

func (m TuiModel) viewReviewConsentSelection() string {
	return viewYesNo(
		"Would you like to review any duplicate images found?",
		reviewSelectionText,
		func() bool {
			return m.wantsReview
		},
	)
}

func (m TuiModel) viewUserReviewSelection() string {
	return viewYesNo(
		"Would you like to keep the duplicate images?",
		userReviewSelectionText,
		func() bool {
			return m.keepDupes
		},
	)
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

	s += fmt.Sprintf("\n\n%s\n", footerText)

	return s
}

func (m TuiModel) viewResults() string {
	s := fmt.Sprintf("\n%s\n\n", resultsHeaderStyle.Render("Hashimg Results"))
	status := m.imgProcessor.Status

	items := []ResultDisplayItem{
		{"Total Images", strconv.Itoa(int(status.TotalImageCount)), resultsTImagesStyle},
		{"Dupes", strconv.Itoa(int(status.DupeImageCount)), resultsDupeStyle},
		{"Cached", strconv.Itoa(int(status.CachedImageCount)), resultsCacheStyle},
		{"New", strconv.Itoa(int(status.NewImageCount)), resultsNewStyle},
		{"", "", resultsValueStyle},
		{"Buffer Size", formatBytes(status.BufferSize), resultsValueStyle},
		{"Analyze Speed", formatDuration(status.AnalyzeTook), resultsValueStyle},
		{"Hash Speed", formatDuration(status.HashingTook), resultsValueStyle},
		{"Filter Speed", formatDuration(status.FilterTook), resultsValueStyle},
		{
			"Update Speed",
			formatDuration(status.UpdatingTook),
			resultsValueStyle,
		},
		{"", "", resultsValueStyle},
		{"Total Time", formatDuration(m.imgProcessor.ProcessTime), resultsTTimeStyle},
	}

	for _, item := range items {
		isInstant := strings.Contains(item.value, "0") &&
			strings.Contains(item.value, "ns") &&
			!strings.Contains(item.value, ".")

		isZeroBytes := item.value == "0 Bytes"

		if isInstant || isZeroBytes {
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

func viewYesNo(question string, header string, isYes func() bool) string {
	s := ""
	if len(header) > 0 {
		s += "\n" + headerStyle.Render(header) + "\n\n"
	}
	s += baseStyle.Foreground(lipgloss.Color(whiteColor)).
		Render(question) +
		"\n\n"
	if !isYes() {
		s += noStyle.Render("> No") + "\n"
		s += baseStyle.Render("  Yes")
	} else {
		s += baseStyle.Render("  No") + "\n"
		s += brightStyle.Render("> Yes")
	}
	s += "\n\n" + footerText + "\n"
	return s
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	suffixes := []string{"Bytes", "KiB", "MiB", "GiB", "TiB"}
	for _, suffix := range suffixes {
		if bytes < 1024 {
			return fmt.Sprintf("%d %s", bytes, timeNotationStyle.Render(suffix))
		}
		bytes /= 1024
	}

	panic("unsupported size when formatting bytes")
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

func getPrettyVersion() string {
	v := internal.GetVersion()
	if strings.Contains(v, "build-") {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6000FF")).Render(v)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#009E7B")).Render(v)
}
