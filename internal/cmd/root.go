package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	project1 string
	project2 string
	format   string
	showAll  bool
)

var rootCmd = &cobra.Command{
	Use:   "gcdiff",
	Short: "Compare and audit GCP resources across projects",
	Long: `gcdiff is a terminal tool for comparing GCP resources across projects.
It helps audit differences between resources, especially useful when
replicating environments or validating Terraform configurations.

Example:
  gcdiff compute my-instance-1 my-instance-2 --project1=prod --project2=staging`,
	Version: "0.3.0",
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gcdiff.yaml)")
	rootCmd.PersistentFlags().StringVar(&project1, "project1", "", "First GCP project ID")
	rootCmd.PersistentFlags().StringVar(&project2, "project2", "", "Second GCP project ID (defaults to project1 if not specified)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "diff", "Output format: diff, json")
	rootCmd.PersistentFlags().BoolVar(&showAll, "show-all", false, "Show all fields including ignored ones")

	// Bind flags to viper
	_ = viper.BindPFlag("project1", rootCmd.PersistentFlags().Lookup("project1"))
	_ = viper.BindPFlag("project2", rootCmd.PersistentFlags().Lookup("project2"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("show-all", rootCmd.PersistentFlags().Lookup("show-all"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not find home directory: %v\n", err)
			return
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gcdiff")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
