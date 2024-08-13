package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/hashimg/internal"
	"github.com/jaeiya/hashimg/internal/models"
	"github.com/jaeiya/hashimg/internal/ui"
)

const appVersion = "1.2.4"

func main() {
	wd, _ := os.Getwd()
	hashPrefix := "0x@"
	iMap, err := internal.MapImages(wd, hashPrefix)
	if err != nil {
		if errors.Is(err, internal.ErrNoImages) {
			fmt.Println(ui.CautionStyle.Render("No images found in current directory"))
			os.Exit(0)
			return
		}
		panic(err)
	}

	workFunc := func(ps *models.ProcessStatus, useAvgBufferSize bool) {
		imgProcessor := internal.NewImageProcessor(hashPrefix, iMap, ps)
		// Errors are handled inside TUI
		_ = imgProcessor.Process(wd, 32, useAvgBufferSize)
	}

	tui := ui.NewTUI(appVersion, workFunc)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
