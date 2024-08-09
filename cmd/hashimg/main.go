package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/go-template/internal/lib"
	"github.com/jaeiya/go-template/internal/lib/models"
	"github.com/jaeiya/go-template/internal/lib/ui"
)

func main() {
	wd, _ := os.Getwd()
	hashPrefix := "0xxs@"
	iMap, err := lib.MapImages(wd, hashPrefix)
	if err != nil {
		panic(err)
	}

	workFunc := func(ps *models.ProcessStatus) {
		imgProcessor := lib.NewImageProcessor(hashPrefix, iMap, ps)
		err = imgProcessor.Process(wd, 32)
		if err != nil {
			panic(err)
		}
	}

	tui := ui.NewTUI(workFunc)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
