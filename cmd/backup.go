package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Didstopia/githubby/ghapi"
	"github.com/spf13/cobra"
	pb "gopkg.in/cheggaaa/pb.v1"
	"gopkg.in/src-d/go-git.v4"
)

// GitHubUser is the GitHub user or organization
var GitHubUser string

// BackupOutputPath is the path where backups are stored
var BackupOutputPath string

// BackupLimit limits the amount of data to backup
var BackupLimit int

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup GitHub repositories",
	Long:  `Backup one or more GitHub repositories`,
	Run: func(cmd *cobra.Command, args []string) {
		// Track progress bar state
		progressEnabled := !Verbose

		// Create a new GitHub client
		client, err := ghapi.NewGitHub(GitHubToken, Verbose)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// Notify the user
		if !Verbose {
			fmt.Println("\nFetching repositories, please wait..")
		}

		// Fetch all repositories for the user
		repositories, err := client.GetRepositories(GitHubUser, BackupLimit)
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

		// Use the current directory as the default output path
		if len(BackupOutputPath) == 0 {
			BackupOutputPath, _ = os.Getwd() // TODO: Add error handling
			if Verbose {
				fmt.Println("Using default backup path:", BackupOutputPath)
			}
		}

		// Loop through repositories and backup/sync them locally
		for _, repository := range repositories {
			// Format the repository details
			repoName := *repository.Name
			repoOwner := *repository.Owner
			repoOwnerName := *repoOwner.Login
			repoURL := "https://github.com/" + repoOwnerName + "/" + repoName

			// Define the final backup location
			backupPath := BackupOutputPath + "/github.com/" + repoOwnerName + "/" + repoName

			if Verbose {
				fmt.Println("Backing up repository from", repoURL, "to", backupPath)
			}

			// Remove the release
			if !DryRun {
				// If an error occurs, we'll simply log it and move on to the next one
				err := client.BackupRepository(GitHubUser, GitHubToken, backupPath, repository, true)
				if err != nil {
					if err == git.NoErrAlreadyUpToDate {
						if Verbose {
							fmt.Println("Repository already up to date:", repoURL)
						}
					} else {
						fmt.Println("Error: Failed to backup repository from", repoURL+":", err)
						// os.Exit(1)
					}
				} else {
					if Verbose {
						fmt.Println("Successfully backed up repository from", repoURL)
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
