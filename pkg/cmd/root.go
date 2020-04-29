package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	githubTokenFlag = "github-token"
	branchNameFlag  = "branch-name"
	fromFlag        = "from"
	toFlag          = "to"
	serviceFlag     = "service"
	cacheDirFlag    = "cache-dir"
	msgFlag         = "commit-message"
	nameFlag        = "commit-name"
	emailFlag       = "commit-email"
	debugFlag       = "debug"
	keepCacheFlag   = "keep-cache"
)

var rootCmd = &cobra.Command{
	Use:   "services",
	Short: "manage services lifecycle via GitOps",
}

func init() {
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
	rootCmd.PersistentFlags().String(
		githubTokenFlag,
		"",
		"oauth access token to authenticate the request",
	)
	logIfError(cobra.MarkFlagRequired(rootCmd.PersistentFlags(), githubTokenFlag))
	logIfError(viper.BindPFlag(githubTokenFlag, rootCmd.PersistentFlags().Lookup(githubTokenFlag)))
	rootCmd.AddCommand(makePromoteCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
