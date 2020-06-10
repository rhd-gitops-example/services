package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	branchNameFlag         = "branch-name"
	cacheDirFlag           = "cache-dir"
	emailFlag              = "commit-email"
	msgFlag                = "commit-message"
	nameFlag               = "commit-name"
	debugFlag              = "debug"
	githubTokenFlag        = "github-token"
	insecureSkipVerifyFlag = "insecure-skip-verify"
	keepCacheFlag          = "keep-cache"
	repoTypeFlag           = "repository-type"
)

var rootCmd = &cobra.Command{
	Use:   "services",
	Short: "manage services lifecycle via GitOps",
}

// Execute is the main entry point into this component.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		postInitCommands(rootCmd.Commands())
	})

	// Required flags
	rootCmd.PersistentFlags().String(
		githubTokenFlag,
		"",
		"oauth access token to authenticate the request",
	)
	logIfError(cobra.MarkFlagRequired(rootCmd.PersistentFlags(), githubTokenFlag))
	logIfError(viper.BindPFlag(githubTokenFlag, rootCmd.PersistentFlags().Lookup(githubTokenFlag)))

	// Optional flags
	rootCmd.PersistentFlags().String(
		branchNameFlag,
		"",
		"the name of the branch on the destination repository for the pull request (auto-generated if empty)",
	)
	logIfError(viper.BindPFlag(branchNameFlag, rootCmd.PersistentFlags().Lookup(branchNameFlag)))

	rootCmd.PersistentFlags().String(
		cacheDirFlag,
		"~/.promotion/cache",
		"where to cache Git checkouts",
	)
	logIfError(viper.BindPFlag(cacheDirFlag, rootCmd.PersistentFlags().Lookup(cacheDirFlag)))

	rootCmd.PersistentFlags().String(
		emailFlag,
		"",
		"the email to use for commits when creating branches",
	)
	logIfError(viper.BindPFlag(emailFlag, rootCmd.PersistentFlags().Lookup(emailFlag)))

	rootCmd.PersistentFlags().String(
		msgFlag,
		"",
		"the msg to use on the resultant commit and pull request",
	)
	logIfError(viper.BindPFlag(msgFlag, rootCmd.PersistentFlags().Lookup(msgFlag)))

	rootCmd.PersistentFlags().String(
		nameFlag,
		"",
		"the name to use for commits when creating branches",
	)
	logIfError(viper.BindPFlag(nameFlag, rootCmd.PersistentFlags().Lookup(nameFlag)))

	rootCmd.PersistentFlags().Bool(
		debugFlag,
		false,
		"additional debug logging output",
	)
	logIfError(viper.BindPFlag(debugFlag, rootCmd.PersistentFlags().Lookup(debugFlag)))

	rootCmd.PersistentFlags().Bool(
		insecureSkipVerifyFlag,
		false,
		"Insecure skip verify TLS certificate",
	)
	logIfError(viper.BindPFlag(insecureSkipVerifyFlag, rootCmd.PersistentFlags().Lookup(insecureSkipVerifyFlag)))

	rootCmd.PersistentFlags().Bool(
		keepCacheFlag,
		false,
		"whether to retain the locally cloned repositories in the cache directory",
	)
	logIfError(viper.BindPFlag(keepCacheFlag, rootCmd.PersistentFlags().Lookup(keepCacheFlag)))

	rootCmd.PersistentFlags().String(
		repoTypeFlag,
		"github",
		"the type of repository: github, gitlab or ghe",
	)
	logIfError(viper.BindPFlag(repoTypeFlag, rootCmd.PersistentFlags().Lookup(repoTypeFlag)))
}

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	logIfError(viper.BindPFlags(cmd.Flags()))
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			logIfError(cmd.Flags().Set(f.Name, viper.GetString(f.Name)))
		}
	})
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
