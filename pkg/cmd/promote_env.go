package cmd

import (
	"fmt"

	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "promote between environment folders within one repository",
	RunE:  promoteEnvAction,
}

func init() {
	promoteCmd.AddCommand(promoteEnvCmd)

	promoteEnvCmd.Flags().String(fromFlag, "", "the source environment folder")
	promoteEnvCmd.Flags().String(toFlag, "", "the destination environment folder")
	promoteEnvCmd.Flags().String(serviceFlag, "", "the name of the service to promote")
	promoteEnvCmd.Flags().String(repoFlag, "", "the URL of the Git repository")

	logIfError(promoteEnvCmd.MarkFlagRequired(fromFlag))
	logIfError(promoteEnvCmd.MarkFlagRequired(toFlag))
	logIfError(promoteEnvCmd.MarkFlagRequired(serviceFlag))
	logIfError(promoteEnvCmd.MarkFlagRequired(repoFlag))
}

func promoteEnvAction(c *cobra.Command, args []string) error {
	bindFlags(c.Flags(), []string{
		fromFlag,
		toFlag,
		serviceFlag,
		repoFlag,
	})

	fromEnvFolder := viper.GetString(fromFlag)
	toEnvFolder := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)
	repo := viper.GetString(repoFlag)

	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	keepCache := viper.GetBool(keepCacheFlag)

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

	sm, err := newServiceManager()
	if err != nil {
		return err
	}

	if msg == "" {
		msg = fmt.Sprintf("Promote environment %s to %s", fromEnvFolder, toEnvFolder)
	}

	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
