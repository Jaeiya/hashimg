package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaeiya/hashimg/internal"
	"github.com/jaeiya/hashimg/internal/ui"
)

const (
	hashPrefix       = "0x@"
	hashLength       = 32
	dupeReviewFolder = "__dupes"
)

func main() {
	wd, _ := os.Getwd()
	iMap, err := internal.MapImages(wd, hashPrefix)
	if err != nil {
		if errors.Is(err, internal.ErrNoImages) {
			fmt.Println(ui.CautionStyle.Render("No images found in current directory"))
			os.Exit(0)
			return
		}
		panic(err)
	}

	imgProcessor := internal.NewImageProcessor(
		internal.ImageProcessorConfig{
			WorkingDir:       wd,
			Prefix:           hashPrefix,
			ImageMap:         iMap,
			HashLength:       hashLength,
			DupeReviewFolder: dupeReviewFolder,
			OpenReviewFolder: true,
		},
	)

	tui := ui.NewTUI(imgProcessor)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
