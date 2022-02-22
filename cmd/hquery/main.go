package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/igolaizola/hquery"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	// Create signal based context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
			cancel()
		}
		signal.Stop(c)
	}()

	// Launch command
	cmd := newCommand()
	if err := cmd.ParseAndRun(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func newCommand() *ffcli.Command {
	fs := flag.NewFlagSet("hquery", flag.ExitOnError)

	return &ffcli.Command{
		ShortUsage: "hquery [flags] <subcommand>",
		FlagSet:    fs,
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			newGetCommand(),
		},
	}
}

func newGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	_ = fs.String("config", "", "config file (optional)")

	url := fs.String("url", "", "url to get doc from")
	file := fs.String("file", "", "file to get doc from")
	query := fs.String("query", "", "doc html query")
	attr := fs.String("attr", "", "html element attribute")

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "hquery get [flags] <key> <value data...>",
		Options: []ff.Option{
			ff.WithConfigFileFlag("config"),
			ff.WithConfigFileParser(ff.PlainParser),
			ff.WithEnvVarPrefix("HQUERY"),
		},
		ShortHelp: "query html data",
		FlagSet:   fs,
		Exec: func(ctx context.Context, args []string) error {
			if *url == "" && *file == "" {
				return errors.New("url or file not provided")
			}
			if *query == "" {
				return errors.New("query not provided")
			}
			text, err := hquery.Get(ctx, *file, *url, *query, *attr)
			if err != nil {
				return err
			}
			fmt.Println(text)
			return nil
		},
	}
}
