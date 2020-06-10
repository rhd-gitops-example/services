package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "promote from one environment folder to another in the same repository",
	RunE:  promoteEnvAction,
}

const (
	repoFlag = "repo"
)

func init() {
	promoteCmd.AddCommand(promoteEnvCmd)

	promoteEnvCmd.Flags().String(
		repoFlag,
		"",
		"Git repository to work within",
	)
	logIfError(cobra.MarkFlagRequired(promoteEnvCmd.Flags(), repoFlag))
	logIfError(viper.BindPFlag(repoFlag, promoteEnvCmd.Flags().Lookup(repoFlag)))
}

func promoteEnvAction(c *cobra.Command, args []string) error {
	// Required flags
	fromEnvFolder := viper.GetString(fromFlag)
	toEnvFolder := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)
	repo := viper.GetString(repoFlag)

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
		RepoPath: repo,
		Branch:   "master",
		Folder:   fromEnvFolder,
	}
	to := promotion.EnvLocation{
		RepoPath: repo,
		Branch:   "master",
		Folder:   toEnvFolder,
	}

	sm := promotion.New(cacheDir, author, promotion.WithDebug(debug), promotion.WithInsecureSkipVerify(insecureSkipVerify), promotion.WithRepoType(repoType))
	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
