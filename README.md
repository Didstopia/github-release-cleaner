[![Build Status](https://travis-ci.org/Didstopia/githubby.svg?branch=master)](https://travis-ci.org/Didstopia/githubby)

# GitHubby

A multi-purpose command line tool for GitHub.

**NOTE:** _Work in progress, not ready for production use!_

### Usage

Backing up repositories:
> githubby --token \<GitHub API Token\> backup --user \<GitHub User or Organization\> --output \<local path for saving repositories\>

Cleaning up releases:
> githubby --token \<GitHub API Token\> clean --repository \<user/repository\> --filter-count \<keep this amount of releases\> --filter-days \<keep releases newer than this\>

### Features (planned vs. implemented)

- [x] Release cleanup (remove releases based on different filters)
- [x] Repo backup and sync (backup one/more/all repositories and sync them)
- [ ] Full backup and sync (same as repo backup, but with support for backing up issues etc.)

### Development

Install/build dependencies:  
> make deps  

Run the application:  
> go run main.go  

Run tests:  
> make test  

### License

See [LICENSE](LICENSE).
