package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "promote between branches with only one environment folder in the same repository",
	RunE:  promoteBranchAction,
}

func init() {
	promoteCmd.AddCommand(promoteBranchCmd)
}

func promoteBranchAction(c *cobra.Command, args []string) error {
	// Required flags
	fromBranch := viper.GetString(fromFlag)
	toBranch := viper.GetString(toFlag)
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
		Branch:   fromBranch,
		Folder:   "",
	}
	to := promotion.EnvLocation{
		RepoPath: repo,
		Branch:   toBranch,
		Folder:   "",
	}

	sm := promotion.New(cacheDir, author, promotion.WithDebug(debug), promotion.WithInsecureSkipVerify(insecureSkipVerify), promotion.WithRepoType(repoType))
	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
