package cmd

import (
	"errors"

	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-gitconfig"
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "promote from one environment to another",
}

// These flags are used in multiple subcommands
const (
	fromFlag    = "from"
	serviceFlag = "service"
	toFlag      = "to"
)

func init() {
	rootCmd.AddCommand(promoteCmd)
}

func newAuthor() (*git.Author, error) {
	name := viper.GetString(nameFlag)
	email := viper.GetString(emailFlag)
	token := viper.GetString(githubTokenFlag)

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
