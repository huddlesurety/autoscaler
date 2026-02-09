package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Run failed", slog.Any("error", err))
	}
}

func run() error {
	fmt.Println("hello")
	return nil
}

func init() {
	lg := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
			TimeFormat: time.Kitchen,
		}),
	)
	slog.SetDefault(lg)
}
