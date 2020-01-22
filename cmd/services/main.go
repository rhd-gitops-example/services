package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bigkevmcd/promote/pkg/cache"
	"github.com/bigkevmcd/promote/pkg/promote"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

const (
	githubTokenFlag = "github-token"
	fromFlag        = "from"
	toFlag          = "to"
	serviceFlag     = "service"
	mappingFileFlag = "mapping-file"
	cacheDirFlag    = "cache-dir"
)

var (
	globalFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     githubTokenFlag,
			Value:    "",
			Usage:    "oauth access token to authenticate the request",
			EnvVars:  []string{"GITHUB_TOKEN"},
			Required: true,
		},
	}

	promoteFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     fromFlag,
			Usage:    "source environment name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     toFlag,
			Usage:    "destination environment name",
			Required: true,
		},
		&cli.StringFlag{
			Name:     serviceFlag,
			Usage:    "service name to promote",
			Required: true,
		},

		&cli.StringFlag{
			Name:     mappingFileFlag,
			Value:    "~/.env-mapping",
			Usage:    "mapping from environment to repository",
			Required: false,
		},

		&cli.StringFlag{
			Name:     cacheDirFlag,
			Value:    "~/.promotion/cache",
			Usage:    "where to cache Git checkouts",
			Required: false,
		},
	}
)

func main() {
	app := &cli.App{
		Name:  "promotion-tool",
		Usage: "promote a file between two git repositories",
		Commands: []*cli.Command{
			{
				Name:   "promote",
				Usage:  "promote from one environment to another",
				Flags:  promoteFlags,
				Action: promoteAction,
			},
		},
		Flags: globalFlags,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func promoteAction(c *cli.Context) error {
	fromRepo := c.String(fromFlag)
	toRepo := c.String(toFlag)
	service := c.String(serviceFlag)
	mappingFilename := c.String(mappingFileFlag)
	token := c.String(githubTokenFlag)
	newBranchName := "testing"

	cacheDir, err := homedir.Expand(c.String(cacheDirFlag))
	if err != nil {
		return fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	cache, err := cache.NewLocalCache(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to create a local cache in %v: %w", cacheDir, err)
	}

	mapping, err := promote.LoadMappingFromFile(mappingFilename)
	if err != nil {
		return fmt.Errorf("failed to load the mappingfile %s: %w", mappingFilename, err)
	}

	return promote.PromoteService(cache, token, service, fromRepo, toRepo, newBranchName, mapping)
}
