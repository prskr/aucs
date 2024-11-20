package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/handlers/cli"
	"github.com/prskr/aucs/infrastructure/config"
)

func main() {
	exitCode := 0
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	if err := run(ctx); err != nil {
		slog.Error("Error occurred", slog.String("err", err.Error()))
		exitCode++
	}
	cancel()

	os.Exit(exitCode)
}

func run(ctx context.Context) error {
	var app App

	kongCtx := kong.Parse(
		&app,
		kong.Name("library"),
		kong.Description("A simple library application for working with messaging"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.BindTo(os.Stdout, (*ports.STDOUT)(nil)),
		kong.Vars{
			"XDG_CACHE_HOME": filepath.ToSlash(xdg.CacheHome),
		},
	)

	return kongCtx.Run()
}

type App struct {
	Logging config.Logging `embed:"" prefix:"logging."`

	Enrich cli.EnrichCLiHandler `cmd:"" help:"Enrich SBOM with available updates" default:"withargs"`
}

func (a *App) AfterApply(kongCtx *kong.Context) error {
	defaultLogger := slog.New(slog.NewJSONHandler(os.Stderr, a.Logging.Options()))

	slog.SetDefault(defaultLogger)
	kongCtx.Bind(defaultLogger)

	return nil
}
