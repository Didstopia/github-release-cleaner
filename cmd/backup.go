package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Didstopia/githubby/ghapi"
	"github.com/spf13/cobra"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// GitHubUser is the GitHub user or organization
var GitHubUser string

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup GitHub repositories",
	Long:  `Backup one or more GitHub repositories`,
	Run: func(cmd *cobra.Command, args []string) {
		// Track progress bar state
		progressEnabled := !Verbose

		// Create a new GitHub client
		client, err := ghapi.NewGitHub(GitHubToken)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Notify the user
		if !Verbose {
			fmt.Println("\nFetching repositories, please wait..")
		}

		// Fetch all repositories for the user
		repositories, err := client.GetRepositories(GitHubUser)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if Verbose {
			fmt.Println("Found", len(repositories), "repositories total")
		}

		if DryRun && Verbose {
			fmt.Println("Dry run detected, simulating backup")
		}

		// Notify the user
		if !Verbose {
			if !DryRun {
				fmt.Printf("Found %d repositories total, starting backup..\n\n", len(repositories))
			} else {
				fmt.Printf("Found %d repositories total, starting simulated backup..\n\n", len(repositories))
			}
		}

		// Create a new progress bar based on the total repository count
		if progressEnabled && len(repositories) > 0 {
			progressBar = pb.StartNew(len(repositories))
		}

		// Loop through repositories and backup/sync them locally
		for _, repository := range repositories {
			if Verbose {
				fmt.Println("Backing up repository", *repository.Name)
			}

			// Remove the release
			if !DryRun {
				// If an error occurs, we'll simply log it and move on to the next one
				err := client.BackupRepository(GitHubUser, repository)
				if err != nil {
					fmt.Println("Error backing up repository", *repository.Name+":", err)
				} else {
					if Verbose {
						fmt.Println("Successfully backed up repository", *repository.Name)
					}
				}
			} else {
				if Verbose {
					fmt.Println("Dry run enabled, simulating backup")
				}
				time.Sleep(time.Duration(250) * time.Millisecond)
			}

			// Increment the progress bar
			if progressEnabled && progressBar != nil {
				progressBar.Increment()
			}
		}

		// Mark the progress bar as done
		if progressEnabled && progressBar != nil {
			progressBar.FinishPrint("\nSuccessfully backed up " + strconv.Itoa(len(repositories)) + " repositories!")
		}
	},
}
