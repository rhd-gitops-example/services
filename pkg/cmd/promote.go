package cmd

import (
	"errors"
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-gitconfig"
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "promote from one environment to another",
	RunE:  promoteAction,
}

const (
	fromFlag          = "from"
	fromBranchFlag    = "from-branch"
	fromEnvFolderFlag = "from-env-folder"
	serviceFlag       = "service"
	toFlag            = "to"
	toBranchFlag      = "to-branch"
	toEnvFolderFlag   = "to-env-folder"
)

func init() {
	rootCmd.AddCommand(promoteCmd)

	// Required flags
	promoteCmd.Flags().String(
		fromFlag,
		"",
		"source Git repository",
	)
	logIfError(promoteCmd.MarkFlagRequired(fromFlag))
	logIfError(viper.BindPFlag(fromFlag, promoteCmd.Flags().Lookup(fromFlag)))

	promoteCmd.Flags().String(
		toFlag,
		"",
		"destination Git repository",
	)
	logIfError(promoteCmd.MarkFlagRequired(toFlag))
	logIfError(viper.BindPFlag(toFlag, promoteCmd.Flags().Lookup(toFlag)))

	promoteCmd.Flags().String(
		serviceFlag,
		"",
		"service name to promote",
	)
	logIfError(cobra.MarkFlagRequired(promoteCmd.Flags(), serviceFlag))
	logIfError(viper.BindPFlag(serviceFlag, promoteCmd.Flags().Lookup(serviceFlag)))

	// Optional flags
	promoteCmd.Flags().String(
		fromBranchFlag,
		"master",
		"branch on the source Git repository",
	)
	logIfError(viper.BindPFlag(fromBranchFlag, promoteCmd.Flags().Lookup(fromBranchFlag)))

	promoteCmd.Flags().String(
		fromEnvFolderFlag,
		"",
		"env folder on the source Git repository (if not provided, the repository should only have one environment)",
	)
	logIfError(viper.BindPFlag(fromEnvFolderFlag, promoteCmd.Flags().Lookup(fromEnvFolderFlag)))

	promoteCmd.Flags().String(
		toBranchFlag,
		"master",
		"branch on the destination Git repository",
	)
	logIfError(viper.BindPFlag(toBranchFlag, promoteCmd.Flags().Lookup(toBranchFlag)))

	promoteCmd.Flags().String(
		toEnvFolderFlag,
		"",
		"env folder on the destination Git repository (if not provided, the repository should only have one environment)",
	)
	logIfError(viper.BindPFlag(toEnvFolderFlag, promoteCmd.Flags().Lookup(toEnvFolderFlag)))
}

func promoteAction(c *cobra.Command, args []string) error {
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
