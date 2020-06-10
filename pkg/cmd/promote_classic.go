package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var promoteComplexCmd = &cobra.Command{
	Use:   "complex",
	Short: "promote arbitrarily from one repo/branch/folder to another",
	RunE:  promoteComplexAction,
}

const (
	fromBranchFlag    = "from-branch"
	fromEnvFolderFlag = "from-env-folder"
	toBranchFlag      = "to-branch"
	toEnvFolderFlag   = "to-env-folder"
)

func init() {
	promoteCmd.AddCommand(promoteComplexCmd)

	// Required flags
	promoteComplexCmd.Flags().String(
		fromFlag,
		"",
		"source Git repository",
	)
	logIfError(promoteComplexCmd.MarkFlagRequired(fromFlag))
	logIfError(viper.BindPFlag(fromFlag, promoteComplexCmd.Flags().Lookup(fromFlag)))

	promoteComplexCmd.Flags().String(
		toFlag,
		"",
		"destination Git repository",
	)
	logIfError(promoteComplexCmd.MarkFlagRequired(toFlag))
	logIfError(viper.BindPFlag(toFlag, promoteComplexCmd.Flags().Lookup(toFlag)))

	promoteComplexCmd.Flags().String(
		serviceFlag,
		"",
		"service name to promote",
	)
	logIfError(cobra.MarkFlagRequired(promoteComplexCmd.Flags(), serviceFlag))
	logIfError(viper.BindPFlag(serviceFlag, promoteComplexCmd.Flags().Lookup(serviceFlag)))

	// Optional flags
	promoteComplexCmd.Flags().String(
		fromBranchFlag,
		"master",
		"branch on the source Git repository",
	)
	logIfError(viper.BindPFlag(fromBranchFlag, promoteComplexCmd.Flags().Lookup(fromBranchFlag)))

	promoteComplexCmd.Flags().String(
		fromEnvFolderFlag,
		"",
		"env folder on the source Git repository (if not provided, the repository should only have one environment)",
	)
	logIfError(viper.BindPFlag(fromEnvFolderFlag, promoteComplexCmd.Flags().Lookup(fromEnvFolderFlag)))

	promoteComplexCmd.Flags().String(
		toBranchFlag,
		"master",
		"branch on the destination Git repository",
	)
	logIfError(viper.BindPFlag(toBranchFlag, promoteComplexCmd.Flags().Lookup(toBranchFlag)))

	promoteComplexCmd.Flags().String(
		toEnvFolderFlag,
		"",
		"env folder on the destination Git repository (if not provided, the repository should only have one environment)",
	)
	logIfError(viper.BindPFlag(toEnvFolderFlag, promoteComplexCmd.Flags().Lookup(toEnvFolderFlag)))
}

func promoteComplexAction(c *cobra.Command, args []string) error {
	// Required flags
	fromRepo := viper.GetString(fromFlag)
	toRepo := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)

	// Optional flags
	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	debug := viper.GetBool(debugFlag)
	fromBranch := viper.GetString(fromBranchFlag)
	fromEnvFolder := viper.GetString(fromEnvFolderFlag)
	insecureSkipVerify := viper.GetBool(insecureSkipVerifyFlag)
	keepCache := viper.GetBool(keepCacheFlag)
	repoType := viper.GetString(repoTypeFlag)
	toBranch := viper.GetString(toBranchFlag)
	toEnvFolder := viper.GetString(toEnvFolderFlag)

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
		Branch:   fromBranch,
		Folder:   fromEnvFolder,
	}
	to := promotion.EnvLocation{
		RepoPath: toRepo,
		Branch:   toBranch,
		Folder:   toEnvFolder,
	}

	sm := promotion.New(cacheDir, author, promotion.WithDebug(debug), promotion.WithInsecureSkipVerify(insecureSkipVerify), promotion.WithRepoType(repoType))
	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}
