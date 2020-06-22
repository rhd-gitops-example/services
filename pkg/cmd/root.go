package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
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

func init() {
	rootCmd.PersistentFlags().String(emailFlag, "", "the email to use for commits when creating branches")
	rootCmd.PersistentFlags().String(msgFlag, "", "the message to use on the resultant commit and pull request")
	rootCmd.PersistentFlags().String(nameFlag, "", "the name to use for commits when creating branches")
	rootCmd.PersistentFlags().Bool(debugFlag, false, "additional debug logging output")
	rootCmd.PersistentFlags().String(githubTokenFlag, "", "oauth access token to authenticate the request")
	rootCmd.PersistentFlags().Bool(insecureSkipVerifyFlag, false, "Insecure skip verify TLS certificate")
	rootCmd.PersistentFlags().String(repoTypeFlag, "github", "the type of repository: github, gitlab or ghe")

	logIfError(cobra.MarkFlagRequired(rootCmd.PersistentFlags(), githubTokenFlag))

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		postInitCommands(rootCmd.Commands())
	})
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

// Execute is the main entry point into this component.
func Execute() {
	bindFlags(rootCmd.PersistentFlags(), []string{
		emailFlag,
		msgFlag,
		nameFlag,
		debugFlag,
		githubTokenFlag,
		insecureSkipVerifyFlag,
		repoTypeFlag,
	})

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func bindFlags(set *pflag.FlagSet, flags []string) {
	for _, f := range flags {
		logIfError(viper.BindPFlag(f, set.Lookup(f)))
	}
}
