package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "promote from a repository with one environment folder to another",
	RunE:  promoteRepoAction,
}

func init() {
	promoteCmd.AddCommand(promoteRepoCmd)

	// Required flags
	promoteRepoCmd.Flags().String(
		fromFlag,
		"",
		"source Git repository",
	)
	logIfError(promoteRepoCmd.MarkFlagRequired(fromFlag))
	logIfError(viper.BindPFlag(fromFlag, promoteRepoCmd.Flags().Lookup(fromFlag)))

	promoteRepoCmd.Flags().String(
		toFlag,
		"",
		"destination Git repository",
	)
	logIfError(promoteRepoCmd.MarkFlagRequired(toFlag))
	logIfError(viper.BindPFlag(toFlag, promoteRepoCmd.Flags().Lookup(toFlag)))

	promoteRepoCmd.Flags().String(
		serviceFlag,
		"",
		"service name to promote",
	)
	logIfError(cobra.MarkFlagRequired(promoteRepoCmd.Flags(), serviceFlag))
	logIfError(viper.BindPFlag(serviceFlag, promoteRepoCmd.Flags().Lookup(serviceFlag)))
}

func promoteRepoAction(c *cobra.Command, args []string) error {
	// Required flags
	fromRepo := viper.GetString(fromFlag)
	toRepo := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)

	// Optional flags
	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	debug := viper.GetBool(debugFlag)
	insecureSkipVerify := viper.GetBool(insecureSkipVerifyFlag)
	keepCache := viper.GetBool(keepCacheFlag)
	repoType := viper.GetString(repoTypeFlag)

	cacheDir, err := homedir.Expand(viper.GetString(cacheDirFlag))
	if err != nil {
		return fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	author, err := newAuthor()
	if err != nil {
		return fmt.Errorf("unable to establish credentials: %w", err)
	}

	from := promotion.EnvLocation{
		RepoPath: fromRepo,
		Branch:   "master",
		Folder:   "",
	}
	to := promotion.EnvLocation{
		RepoPath: toRepo,
		Branch:   "master",
		Folder:   "",
	}

	sm := promotion.New(cacheDir, author, promotion.WithDebug(debug), promotion.WithInsecureSkipVerify(insecureSkipVerify), promotion.WithRepoType(repoType))
	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
