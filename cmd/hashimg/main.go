package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jaeiya/go-template/internal/lib"
	"github.com/jaeiya/go-template/internal/lib/models"
	"github.com/jaeiya/go-template/internal/lib/ui"
)

func main() {
	wd, _ := os.Getwd()
	hashPrefix := "0x@"
	iMap, err := lib.MapImages(wd, hashPrefix)
	if err != nil {
		if errors.Is(err, lib.ErrNoImages) {
			style := lipgloss.NewStyle().
				MarginTop(2).
				MarginLeft(4).
				Padding(1, 2).
				Background(lipgloss.Color("#101010")).
				Foreground(lipgloss.Color("#FFFF5C"))

			fmt.Println(style.Render("No images found in current directory"))
			os.Exit(0)
			return
		}
		panic(err)
	}

	workFunc := func(ps *models.ProcessStatus) {
		imgProcessor := lib.NewImageProcessor(hashPrefix, iMap, ps)
		// Errors are handled inside TUI
		_ = imgProcessor.Process(wd, 32)
	}

	tui := ui.NewTUI(workFunc)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
