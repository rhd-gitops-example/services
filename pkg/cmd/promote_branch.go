package cmd

import (
	"fmt"

	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteBranchCmd = &cobra.Command{
	Use:   "branch",
	Short: "promote between branches within one repository",
	RunE:  promoteBranchAction,
}

func init() {
	promoteCmd.AddCommand(promoteBranchCmd)

	promoteBranchCmd.Flags().String(fromFlag, "", "the source branch")
	promoteBranchCmd.Flags().String(toFlag, "", "the destination branch")
	promoteBranchCmd.Flags().String(serviceFlag, "", "the name of the service to promote")
	promoteBranchCmd.Flags().String(repoFlag, "", "the URL of the Git repository")

	logIfError(promoteBranchCmd.MarkFlagRequired(fromFlag))
	logIfError(promoteBranchCmd.MarkFlagRequired(toFlag))
	logIfError(promoteBranchCmd.MarkFlagRequired(serviceFlag))
	logIfError(promoteBranchCmd.MarkFlagRequired(repoFlag))
}

func promoteBranchAction(c *cobra.Command, args []string) error {
	bindFlags(c.Flags(), []string{
		fromFlag,
		toFlag,
		serviceFlag,
		repoFlag,
	})

	fromBranch := viper.GetString(fromFlag)
	toBranch := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)
	repo := viper.GetString(repoFlag)

	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	keepCache := viper.GetBool(keepCacheFlag)

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

	sm, err := newServiceManager()
	if err != nil {
		return err
	}

	if msg == "" {
		msg = fmt.Sprintf("Promote branch %s to %s", fromBranch, toBranch)
	}

	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
