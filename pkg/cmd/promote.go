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
	branchNameFlag    = "branch-name"
	fromFlag          = "from"
	fromBranchFlag    = "from-branch"
	fromEnvFolderFlag = "from-env-folder"
	serviceFlag       = "service"
	toFlag            = "to"
	toBranchFlag      = "to-branch"
	toEnvFolderFlag   = "to-env-folder"

	repoFlag = "repo" // used by subcommands
)

func init() {
	rootCmd.AddCommand(promoteCmd)

	promoteCmd.PersistentFlags().String(branchNameFlag, "", "the branch on the destination repository for the pull request (auto-generated if empty)")
	promoteCmd.PersistentFlags().String(cacheDirFlag, "~/.promotion/cache", "where to cache Git checkouts")
	promoteCmd.PersistentFlags().Bool(keepCacheFlag, false, "whether to retain the locally cloned repositories in the cache directory")

	promoteCmd.Flags().String(fromFlag, "", "the source Git repository (URL or local)")
	promoteCmd.Flags().String(toFlag, "", "the destination Git repository")
	promoteCmd.Flags().String(serviceFlag, "", "the name of the service to promote")
	promoteCmd.Flags().String(fromBranchFlag, "master", "the branch on the source Git repository")
	promoteCmd.Flags().String(fromEnvFolderFlag, "", "env folder on the source Git repository (if not provided, the repository should only have one folder under environments/)")
	promoteCmd.Flags().String(toBranchFlag, "master", "the branch on the destination Git repository")
	promoteCmd.Flags().String(toEnvFolderFlag, "", "env folder on the destination Git repository (if not provided, the repository should only have one folder under environments/)")

	logIfError(promoteCmd.MarkFlagRequired(fromFlag))
	logIfError(promoteCmd.MarkFlagRequired(toFlag))
	logIfError(promoteCmd.MarkFlagRequired(serviceFlag))
}

func promoteAction(c *cobra.Command, args []string) error {
	bindFlags(c.PersistentFlags(), []string{
		branchNameFlag,
		cacheDirFlag,
		keepCacheFlag,
	})
	bindFlags(c.Flags(), []string{
		fromFlag,
		toFlag,
		serviceFlag,
		fromBranchFlag,
		fromEnvFolderFlag,
		toBranchFlag,
		toEnvFolderFlag,
	})

	// Required flags
	fromRepo := viper.GetString(fromFlag)
	toRepo := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)

	// Optional flags
	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	fromBranch := viper.GetString(fromBranchFlag)
	fromEnvFolder := viper.GetString(fromEnvFolderFlag)
	keepCache := viper.GetBool(keepCacheFlag)
	toBranch := viper.GetString(toBranchFlag)
	toEnvFolder := viper.GetString(toEnvFolderFlag)

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

	sm, err := newServiceManager()
	if err != nil {
		return err
	}

	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}

func newServiceManager() (*promotion.ServiceManager, error) {
	cacheDir, err := homedir.Expand(viper.GetString(cacheDirFlag))
	if err != nil {
		return nil, fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	author, err := newAuthor()
	if err != nil {
		return nil, fmt.Errorf("unable to establish credentials: %w", err)
	}

	return promotion.New(
		cacheDir,
		author,
		promotion.WithDebug(viper.GetBool(debugFlag)),
		promotion.WithInsecureSkipVerify(viper.GetBool(insecureSkipVerifyFlag)),
		promotion.WithRepoType(viper.GetString(repoTypeFlag)),
	), nil
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
