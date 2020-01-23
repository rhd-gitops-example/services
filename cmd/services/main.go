package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bigkevmcd/promote/pkg/cache"
	"github.com/bigkevmcd/promote/pkg/promote"
	"github.com/mitchellh/go-homedir"
	"github.com/tcnksm/go-gitconfig"
	"github.com/urfave/cli/v2"
)

const (
	githubTokenFlag = "github-token"
	fromFlag        = "from"
	toFlag          = "to"
	serviceFlag     = "service"
	mappingFileFlag = "mapping-file"
	cacheDirFlag    = "cache-dir"
	nameFlag        = "commit-name"
	emailFlag       = "commit-email"
)

var (
	globalFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     githubTokenFlag,
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

		&cli.StringFlag{
			Name:     nameFlag,
			Usage:    "the name to use for commits when creating branches",
			Required: false,
			EnvVars:  []string{"COMMIT_NAME"},
		},
		&cli.StringFlag{
			Name:     emailFlag,
			Usage:    "the email to use for commits when creating branches",
			Required: false,
			EnvVars:  []string{"COMMIT_EMAIL"},
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
		log.Fatalf("%v", err)
	}
}

func promoteAction(c *cli.Context) error {
	fromRepo := c.String(fromFlag)
	toRepo := c.String(toFlag)
	service := c.String(serviceFlag)
	token := c.String(githubTokenFlag)
	newBranchName := "testing"

	cacheDir, err := homedir.Expand(c.String(cacheDirFlag))
	if err != nil {
		return fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	user, email, err := credentials(c)
	if err != nil {
		return fmt.Errorf("unable to establish: %w", err)
	}
	cache, err := cache.NewLocalCache(cacheDir, user, email)
	if err != nil {
		return fmt.Errorf("failed to create a local cache in %v: %w", cacheDir, err)
	}

	mappingFilename, err := homedir.Expand(c.String(mappingFileFlag))
	if err != nil {
		return fmt.Errorf("failed to expand mapping file path: %w", err)
	}
	mapping, err := promote.LoadMappingFromFile(mappingFilename)
	if err != nil {
		return fmt.Errorf("failed to load the mappingfile %s: %w", mappingFilename, err)
	}

	return promote.PromoteService(cache, token, service, fromRepo, toRepo, newBranchName, mapping)
}

func credentials(c *cli.Context) (string, string, error) {
	name := c.String(nameFlag)
	email := c.String(emailFlag)

	var err error
	if name == "" {
		name, err = gitconfig.Username()
		if err != nil {
			return "", "", err
		}
	}

	if email == "" {
		email, err = gitconfig.Email()
		if err != nil {
			return "", "", err
		}
	}

	// TODO: make this a multierror with both errors?
	if name == "" || email == "" {
		return "", "", errors.New("unable to identify user and email for commits")
	}

	return name, email, nil
}
