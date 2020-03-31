package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/tcnksm/go-gitconfig"
	"github.com/urfave/cli/v2"

	"github.com/rhd-gitops-example/services/pkg/avancement"
	"github.com/rhd-gitops-example/services/pkg/git"
)

const (
	githubTokenFlag = "github-token"
	branchNameFlag  = "branch-name"
	fromFlag        = "from"
	toFlag          = "to"
	serviceFlag     = "service"
	cacheDirFlag    = "cache-dir"
	nameFlag        = "commit-name"
	emailFlag       = "commit-email"
	debugFlag       = "debug"
	keepCacheFlag   = "keep-cache"
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
			Usage:    "source Git repository",
			Required: true,
		},
		&cli.StringFlag{
			Name:     toFlag,
			Usage:    "destination Git repository",
			Required: true,
		},
		&cli.StringFlag{
			Name:     serviceFlag,
			Usage:    "service name to promote",
			Required: true,
		},
		&cli.StringFlag{
			Name:  branchNameFlag,
			Usage: "the name to use for the newly created branch",
			Value: "test-branch",
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
		&cli.BoolFlag{
			Name:     debugFlag,
			Usage:    "additional debug logging output",
			EnvVars:  []string{"DEBUG_SERVICES"},
			Value:    false,
			Required: false,
		},
		&cli.BoolFlag{
			Name:     keepCacheFlag,
			Usage:    "whether to retain the locally cloned repositories in the cache directory",
			Value:    false,
			Required: false,
		},
	}
)

func main() {
	app := &cli.App{
		Name:        "services",
		Description: "manage services lifecycle via GitOps",
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
	newBranchName := c.String(branchNameFlag)
	debug := c.Bool(debugFlag)
	keepCache := c.Bool(keepCacheFlag)

	cacheDir, err := homedir.Expand(c.String(cacheDirFlag))
	if err != nil {
		return fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	author, err := newAuthor(c)
	if err != nil {
		return fmt.Errorf("unable to establish credentials: %w", err)
	}
	return avancement.New(cacheDir, author, avancement.WithDebug(debug)).Promote(service, fromRepo, toRepo, newBranchName, keepCache)
}

func newAuthor(c *cli.Context) (*git.Author, error) {
	name := c.String(nameFlag)
	email := c.String(emailFlag)
	token := c.String(githubTokenFlag)

	var err error
	if name == "" {
		name, err = gitconfig.Username()
		if err != nil {
			return nil, err
		}
	}

	if email == "" {
		email, err = gitconfig.Email()
		if err != nil {
			return nil, err
		}
	}

	// TODO: make this a multierror with both errors?
	if name == "" || email == "" {
		return nil, errors.New("unable to identify user and email for commits")
	}

	return &git.Author{Name: name, Email: email, Token: token}, nil
}
