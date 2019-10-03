// Package ghapi provides a wrapper for easier access to specific parts the GitHub API.
package ghapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
)

// DefaultResultsPerPage sets the default amount of results to fetch from certain GitHub API endpoints
const DefaultResultsPerPage = 100

// GitHub is an abstraction for the real GitHub API client
type GitHub struct {
	ctx     context.Context
	client  *github.Client
	verbose bool
}

// NewGitHub creates and returns a reference to a new GitHub object
func NewGitHub(token string, verbose bool) (*GitHub, error) {
	githubClient := &GitHub{}

	// FIXME: Validate the token using the GitHub client

	// Create an authentication context for the GitHub API client
	githubClient.ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(githubClient.ctx, ts)

	// Create the actual GitHub API client
	githubClient.client = github.NewClient(tc)

	// Store the verbosity flag
	githubClient.verbose = verbose // FIXME: Actually use verbosity to print stuff wherever possible..

	return githubClient, nil
}

// GetRepositories returns all repositories for the supplied user or organization
func (githubClient *GitHub) GetRepositories(owner string, limit int) ([]*github.Repository, error) {
	// FIXME: Validate the owner using the GitHub client

	repositories, err := githubClient.getAllRepositories(owner, 1, limit, nil)
	if err != nil {
		return nil, err
	}

	// log.Println("Got", len(repositories), "repositories total")

	return repositories, nil
}

// BackupRepository will attempt to backup or sync a repository from GitHub
func (githubClient *GitHub) BackupRepository(owner string, token string, path string, repository *github.Repository, retry bool) error {
	// Create the destination path if it doesn't yet exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}

	// FIXME: We should probably implement graceful shutdown at some point, as otherwise repos can end up in a dirty state..

	// FIXME: This seems to break with submodules, no idea why, so for now we'll just use token + ssh (user needs to have ssh auth setup)
	// Create a basic authentication wrapper (repositories need to use https, not git or ssh)
	// auth := http.BasicAuth{
	// 	Username: owner,
	// 	Password: token,
	// }

	// Setup the git action logging
	logHandler := log.Writer()
	if !githubClient.verbose {
		logHandler = nil
	}

	// Attempt to open an existing repository
	if githubClient.verbose {
		fmt.Println("Checking if a local repository exists for", *repository.Name)
	}
	localRepository, err := git.PlainOpen(path)
	if err != nil {
		// Local repository is invalid or doesn't exist, continue by attempting to clone a fresh copy
		if githubClient.verbose {
			fmt.Println("No existing repository exists, cloning", *repository.Name)
		}
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			// Auth:              &auth,
			// URL:               repository.GetCloneURL(),
			URL: repository.GetGitURL(),
			// URL:               repository.GetSSHURL(),
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Progress:          logHandler,
		})
		if err != nil {
			// Nuke and retry
			if retry {
				if githubClient.verbose {
					fmt.Println("Cloning failed for repository", *repository.Name, "(retrying):", err)
				}
				os.RemoveAll(path)
				return githubClient.BackupRepository(owner, token, path, repository, false)
			}
			if githubClient.verbose {
				fmt.Println("Cloning failed for repository", *repository.Name+":", err)
			}
			return err
		}
	} else {
		// Local repository exists, attempt to get the working tree for it
		if githubClient.verbose {
			fmt.Println("Getting state for repository", *repository.Name)
		}
		worktree, err := localRepository.Worktree()
		if err != nil {
			// Nuke and retry
			if retry {
				if githubClient.verbose {
					fmt.Println("Getting state failed for repository", *repository.Name, "(retrying):", err)
				}
				os.RemoveAll(path)
				return githubClient.BackupRepository(owner, token, path, repository, false)
			}
			if githubClient.verbose {
				fmt.Println("Getting state failed for repository", *repository.Name+":", err)
			}
			return err
		}

		// Attempt to reset the repository
		if githubClient.verbose {
			fmt.Println("Resetting state for repository", *repository.Name)
		}
		if err := worktree.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
			// Nuke and retry
			if retry {
				if githubClient.verbose {
					fmt.Println("Resetting state failed for repository", *repository.Name, "(retrying):", err)
				}
				os.RemoveAll(path)
				return githubClient.BackupRepository(owner, token, path, repository, false)
			}
			if githubClient.verbose {
				fmt.Println("Resetting state failed for repository", *repository.Name+":", err)
			}
			return err
		}

		// Attempt to checkout the latest changes
		if githubClient.verbose {
			fmt.Println("Pulling changesfor repository", *repository.Name)
		}
		if err := worktree.Pull(&git.PullOptions{
			// Auth:     &auth,
			Force:    true,
			Progress: logHandler,
		}); err != nil {
			// Nuke and retry
			if retry && err != git.NoErrAlreadyUpToDate {
				if githubClient.verbose {
					fmt.Println("Pulling changes failed for repository", *repository.Name, "(retrying):", err)
				}
				os.RemoveAll(path)
				return githubClient.BackupRepository(owner, token, path, repository, false)
			}
			if githubClient.verbose {
				fmt.Println("Pulling changes failed for repository", *repository.Name+":", err)
			}
			return err
		}
	}

	// Return nil on success
	return nil
}

// GetReleases returns all release information for the supplied repository
func (githubClient *GitHub) GetReleases(owner string, repository string) ([]*github.RepositoryRelease, error) {
	// FIXME: Validate the owner using the GitHub client
	// FIXME: Validate the repository using the GitHub client

	// Find all releases (handles pagination behind the scenes, starting at page 1)
	releases, err := githubClient.getAllReleases(owner, repository, 1, nil)
	if err != nil {
		return nil, err
	}

	// log.Println("Got", len(releases), "releases total")

	return releases, nil
}

// RemoveRelease will attempt to delete a release from GitHub
func (githubClient *GitHub) RemoveRelease(owner string, repo string, release *github.RepositoryRelease) error {
	// Delete the release
	deleteReleaseErr := githubClient.deleteRelease(release)
	if deleteReleaseErr != nil {
		return deleteReleaseErr
	}

	// Delete the tag
	deleteTagErr := githubClient.deleteTag(owner, repo, release)
	if deleteTagErr != nil {
		return deleteTagErr
	}

	// Return nil on success
	return nil
}

func (githubClient *GitHub) deleteRelease(release *github.RepositoryRelease) error {
	//log.Println("Deleting release:", release.TagName)

	// Create the release deletion request
	req, reqErr := githubClient.client.NewRequest("DELETE", *release.URL, nil)
	if reqErr != nil {
		return reqErr
	}

	// Run the request
	_, doErr := githubClient.client.Do(githubClient.ctx, req, nil)
	if doErr != nil {
		return doErr
	}

	//log.Println("Delete release response:", res)

	// Return nil on success
	return nil
}

func (githubClient *GitHub) deleteTag(owner string, repo string, release *github.RepositoryRelease) error {
	// Construct the API endpoint url
	url := "https://api.github.com/repos/" + owner + "/" + repo + "/git/refs/tags/" + *release.TagName

	//log.Println("Deleting tag:", url)

	// Create the tag deletion request
	req, reqErr := githubClient.client.NewRequest("DELETE", url, nil)
	if reqErr != nil {
		return reqErr
	}

	// Run the request
	_, doErr := githubClient.client.Do(githubClient.ctx, req, nil)
	if doErr != nil {
		return doErr
	}

	// log.Println("Delete tag response:", res)

	// Return nil on success
	return nil
}

func (githubClient *GitHub) getAllRepositories(owner string, page int, limit int, existingRepositories []*github.Repository) ([]*github.Repository, error) {
	// log.Println("Getting repositories for page", page, "with a limit of", limit)

	// Create an array that will eventually contain all repositories
	allRepositories := make([]*github.Repository, 0)

	// Use existing repositories if necessary
	if existingRepositories != nil {
		allRepositories = existingRepositories
	}

	// Set the results per page (DefaultResultsPerPage is set to the maximum allowed value)
	perPage := DefaultResultsPerPage

	// Limit the results per page, if a limit is curently set
	if limit > 0 {
		perPage = limit
	}

	// Calculate the total amount of results and adjust for negative values
	total := len(allRepositories)
	if total <= 0 {
		total = perPage
	}

	// Calculate the amount of results left and adjust for negative values
	left := limit - total
	if limit <= 0 {
		left = total
	}

	// Get repositories for the current page
	opts := &github.RepositoryListOptions{}
	opts.Page = page
	opts.PerPage = perPage
	repositories, res, err := githubClient.client.Repositories.List(githubClient.ctx, owner, opts)
	if err != nil {
		return nil, err
	}

	// Add the current repositories
	for _, repository := range repositories {
		allRepositories = append(allRepositories, repository)
	}

	// Recursively move to the next page if there are any more pages left
	if res.NextPage > 0 && res.NextPage > page && (perPage == DefaultResultsPerPage || (perPage < DefaultResultsPerPage && left > 0)) {
		log.Println("Moving from page", page, "to", res.NextPage)
		return githubClient.getAllRepositories(owner, res.NextPage, limit, allRepositories)
	}

	// Return an error if we have no repositories
	if allRepositories == nil || len(allRepositories) <= 0 {
		return nil, errors.New("no repositories found")
	}

	// Return all repositories if we're done
	return allRepositories, nil
}

func (githubClient *GitHub) getAllReleases(owner string, repository string, page int, existingReleases []*github.RepositoryRelease) ([]*github.RepositoryRelease, error) {
	// log.Println("Getting releases for page ", page)

	// Create an array that will eventually contain all releases
	allReleases := make([]*github.RepositoryRelease, 0)

	// Use existing releases if necessary
	if existingReleases != nil {
		allReleases = existingReleases
	}

	// Get releases for the current page
	releases, res, err := githubClient.client.Repositories.ListReleases(githubClient.ctx, owner, repository, &github.ListOptions{Page: page, PerPage: DefaultResultsPerPage})
	if err != nil {
		return nil, err
	}

	// Add the current releases
	for _, release := range releases {
		allReleases = append(allReleases, release)
	}

	// Recursively move to the next page if there are any more pages left
	if res.NextPage > 0 && res.NextPage > page {
		// log.Println("Moving from page", page, "to", res.NextPage)
		return githubClient.getAllReleases(owner, repository, res.NextPage, allReleases)
	}

	// Return an error if we have no releases
	if allReleases == nil || len(allReleases) <= 0 {
		return nil, errors.New("no releases found")
	}

	// Return all releases if we're done
	return allReleases, nil
}
