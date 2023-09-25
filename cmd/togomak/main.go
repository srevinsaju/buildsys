package main

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/srevinsaju/togomak/v1/pkg/behavior"
	"github.com/srevinsaju/togomak/v1/pkg/cache"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/filter"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/orchestra"
	"github.com/srevinsaju/togomak/v1/pkg/path"
	"github.com/urfave/cli/v2"
	"os"
)

var verboseCount = 0

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	meta.AppVersion = version
	app := cli.NewApp()
	app.Name = meta.AppName
	app.Description = meta.AppDescription
	app.Action = run
	app.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)

	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Usage:   "initialize a new pipeline",
			Aliases: []string{"i"},
			Action:  initPipeline,
		},
		{
			Name:   "run",
			Usage:  "run a pipeline",
			Action: run,
		},
		{
			Name:    "list",
			Usage:   "list all the pipelines",
			Aliases: []string{"ls", "l"},
			Action:  list,
		},
		{
			Name:   "fmt",
			Usage:  "format a pipeline file",
			Action: format,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "check",
					Usage:   "check if the file is formatted",
					Aliases: []string{"c"},
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Usage:   "format all the files named togomak.hcl in the current directory, and its children",
					Aliases: []string{"r"},
				},
			},
		},
		{
			Name:  "cache",
			Usage: "manage the cache",
			Subcommands: []*cli.Command{
				{
					Name:   "clean",
					Usage:  "clean the cache",
					Action: cleanCache,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "recursive",
							Usage:   "clean the cache recursively",
							Aliases: []string{"r"},
						},
					},
				},
			},
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "path to the pipeline file",
		},
		&cli.BoolFlag{
			Name:   "child",
			Usage:  "run the pipeline as a child process (advanced)",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   "parent",
			Usage:  "the parent process id (advanced)",
			Hidden: true,
		},
		&cli.StringSliceFlag{
			Name:   "parent-param",
			Usage:  "parameter passed to child togomak process (advanced)",
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:    "unattended",
			Aliases: []string{"no-prompt", "no-interactive"},
			Usage:   "do not prompt for responses, or wait for user responses. run in auto-pilot",
			EnvVars: []string{"TOGOMAK_UNATTENDED"},
			Value:   !isatty.IsTerminal(os.Stdin.Fd()),
		},
		&cli.BoolFlag{
			Name:    "ci",
			Usage:   "run in CI mode",
			EnvVars: []string{"CI", "TOGOMAK_CI"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"C", "directory"},
			Usage:   "path to the directory where the pipeline file is located",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "enable verbose logging",
			Count:   &verboseCount,
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n", "just-print", "recon"},
			Usage:   "Don't actually run any stage; just print the commands",
			EnvVars: []string{"TOGOMAK_DRY_RUN"},
		},
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "display the version of the application",
	}
	app.Version = meta.AppVersion
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func initPipeline(ctx *cli.Context) error {
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	if dir == "" {
		dir = owd
	}
	orchestra.InitPipeline(dir)
	return nil
}

func newConfigFromCliContext(ctx *cli.Context) conductor.Config {
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	pipelineFilePath := ctx.String("file")
	if pipelineFilePath == "" {
		dir := dir
		if dir == "" {
			dir = owd
		}
		pipelineFilePath = autoDetectFilePath(dir)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	var stages []filter.Item
	for _, stage := range ctx.Args().Slice() {
		stages = append(stages, filter.NewFilterItem(stage))
	}
	cfg := conductor.Config{

		Behavior: &behavior.Behavior{
			Unattended: ctx.Bool("unattended") || ctx.Bool("ci"),
			Ci:         ctx.Bool("ci"),
			DryRun:     ctx.Bool("dry-run"),

			Child: behavior.Child{
				Enabled:      ctx.Bool("child"),
				Parent:       ctx.String("parent"),
				ParentParams: ctx.StringSlice("parent-param"),
			},
		},

		Paths: &path.Path{
			Pipeline: pipelineFilePath,
			Cwd:      dir,
			Owd:      owd,
		},
		User:      os.Getenv("USER"),
		Hostname:  hostname,
		Interface: conductor.Interface{Verbosity: verboseCount},
		Pipeline: conductor.ConfigPipeline{
			Filtered: stages,

			DryRun: ctx.Bool("dry-run"),
		},
	}
	return cfg
}

func run(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)

	t := conductor.NewTogomak(cfg)
	v := orchestra.Perform(t)
	t.Destroy()
	os.Exit(v)
	return nil
}

func cleanCache(ctx *cli.Context) error {
	recursive := ctx.Bool("recursive")
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	if dir == "" {
		dir = owd
	}
	cache.CleanCache(dir, recursive)
	return nil
}

func list(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	return orchestra.List(cfg)
}

func format(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	return orchestra.Format(cfg, ctx.Bool("check"), ctx.Bool("recursive"))
}
