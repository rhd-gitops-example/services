package cmd

import (
	"fmt"

	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "promote between repositories",
	RunE:  promoteRepoAction,
}

func init() {
	promoteCmd.AddCommand(promoteRepoCmd)

	promoteRepoCmd.Flags().String(fromFlag, "", "the URL of the source repository")
	promoteRepoCmd.Flags().String(toFlag, "", "the URL of the destination repository")
	promoteRepoCmd.Flags().String(serviceFlag, "", "the name of the service to promote")

	logIfError(promoteRepoCmd.MarkFlagRequired(fromFlag))
	logIfError(promoteRepoCmd.MarkFlagRequired(toFlag))
	logIfError(promoteRepoCmd.MarkFlagRequired(serviceFlag))
}

func promoteRepoAction(c *cobra.Command, args []string) error {
	bindFlags(c.Flags(), []string{
		fromFlag,
		toFlag,
		serviceFlag,
	})

	fromRepo := viper.GetString(fromFlag)
	toRepo := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)

	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	keepCache := viper.GetBool(keepCacheFlag)

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

	sm, err := newServiceManager()
	if err != nil {
		return err
	}

	if msg == "" {
		msg = fmt.Sprintf("Promote repository %s to %s", fromRepo, toRepo)
	}

	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
