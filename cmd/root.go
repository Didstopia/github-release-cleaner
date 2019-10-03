// Package cmd is the primary entrypoint, and handles command parsing and execution.
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra" // Include the Cobra Commander package
	pb "gopkg.in/cheggaaa/pb.v1"
)

// Verbose can be toggled on/off to enable diagnostic log output
var Verbose bool

// DryRun will simulate the cleanup process without actually deleting anything
var DryRun bool

// GitHubToken is the GitHub API token
var GitHubToken string

// The primary logger
var log = logrus.New()

// The progress bar (only used when running non-verbosely)
var progressBar *pb.ProgressBar

// The primary cobra command object
var rootCmd *cobra.Command

func init() {
	// Create the primary command object
	rootCmd = &cobra.Command{
		Use:   "githubby",
		Short: "GitHub CLI utility",
		Long:  `A multi-purpose CLI utility for interacting with GitHub`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}

	// Add the clean command
	rootCmd.AddCommand(cleanCmd)

	// Add the backup command
	rootCmd.AddCommand(backupCmd)

	// FIXME: This is persisted to config, so can't be easily disabled
	// Add the "verbose" flag globally, so it's available for all commands
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")

	// FIXME: This is persisted to config, so can't be easily disabled
	// Add the "dry-run" flag globally, so it's available for all commands
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dry-run", "D", false, "Simulate running")

	// Add the "token" flag globally, and mark it as always required
	rootCmd.PersistentFlags().StringVarP(&GitHubToken, "token", "t", "", "GitHub API Token (required)")
	rootCmd.MarkPersistentFlagRequired("token")

	// Add the "repository" flag to the clean command and mark it as always required
	cleanCmd.Flags().StringVarP(&GitHubRepository, "repository", "r", "", "GitHub Repository (required, short format only, eg. user/repo)")
	cleanCmd.MarkFlagRequired("repository")

	// Add the "filter-days" flag to the clean command
	cleanCmd.Flags().Int64VarP(&FilterDays, "filter-days", "d", -1, "Filter based on maximum days since release (at least one filter is required)")

	// Add the "filter-count" flag to the clean command
	cleanCmd.Flags().Int64VarP(&FilterCount, "filter-count", "c", -1, "Filter to cleanup releases over the set amount (at least one filter is required)")

	// Add the "user" flag to the backup command
	backupCmd.Flags().StringVarP(&GitHubUser, "user", "u", "", "GitHub user or organization (required)")
	backupCmd.MarkFlagRequired("user")

	// Add the "output" flag to the backup command
	backupCmd.Flags().StringVarP(&BackupOutputPath, "output", "o", "", "Backup output path (defaults to current directory)")

	// Add the "limit" flag to the backup command
	backupCmd.Flags().IntVarP(&BackupLimit, "limit", "l", 0, "Limit the amount of backups (no limit by default)")
}

// Execute starts the Cobra commander, which in turn will handle execution and any arguments
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
