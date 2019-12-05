package cmd

import (
	"fmt"
	"os"
	"time"

	publish "github.com/object88/gha-publish-release-assets"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	excludeKey string = "exclude"
	includeKey        = "include"
)

const (
	authHeaderEnv       string = "AUTH_HEADER"
	githubEventPathEnv         = "GITHUB_EVENT_PATH"
	githubRepositoryEnv        = "GITHUB_REPOSITORY"
	githubTokenEnv             = "GITHUB_TOKEN"
	githubWorkspaceEnv         = "GITHUB_WORKSPACE"
)

// InitializeCommands sets up the cobra commands
func InitializeCommands() *cobra.Command {
	rootCmd := createRootCommand()

	return rootCmd
}

type command struct {
	cobra.Command
	p *publish.Publisher
}

func createRootCommand() *cobra.Command {
	var start time.Time
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "publish",
			Short: "publish adds one or more assets to a Github release",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				start = time.Now()
				return c.Preexecute(cmd, args)
			},
			Run: func(cmd *cobra.Command, args []string) {
				c.Execute(cmd, args)
				cmd.HelpFunc()(cmd, args)
			},
			PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
				duration := time.Since(start)
				fmt.Fprintf(os.Stderr, "Executed command in %s\n", duration)
				return nil
			},
		},
	}

	flags := c.PersistentFlags()
	flags.StringSliceP(excludeKey, excludeKey[0:1], []string{}, "file globs to exclude")

	flags.StringSliceP(includeKey, includeKey[0:1], []string{}, "file globs to include")

	return &c.Command
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	req, err := publish.NewRequest("https://uploads.github.com")
	if err != nil {
		return err
	}

	p := publish.NewPublisher(req)
	c.p = p

	var ok bool

	c.p.Github.Auth, ok = os.LookupEnv(authHeaderEnv)
	if !ok {
		return errors.Errorf("Missing '%s' environment variable; cannot proceed", authHeaderEnv)
	}

	file, ok := os.LookupEnv(githubEventPathEnv)
	if !ok {
		return errors.Errorf("Missing '%s' environment variable; cannot proceed", githubEventPathEnv)
	}
	ge, err := publish.NewGithubEvent(file)
	if err != nil {
		return errors.Wrapf(err, "Failed to read Github event")
	}

	c.p.Github.ReleaseID, err = ge.ReleaseID()
	if err != nil {
		return err
	}

	c.p.Github.Repository, ok = os.LookupEnv(githubRepositoryEnv)
	if !ok {
		return errors.Errorf("Missing '%s' environment varianble; cannot proceed", githubRepositoryEnv)
	}

	c.p.Github.Token, ok = os.LookupEnv(githubTokenEnv)
	if !ok {
		return errors.Errorf("Missing '%s' environment varianble; cannot proceed", githubTokenEnv)
	}

	c.p.Github.Workspace, ok = os.LookupEnv(githubWorkspaceEnv)
	if !ok {
		return errors.Errorf("Missing '%s' environment varianble; cannot proceed", githubWorkspaceEnv)
	}

	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	return c.p.Publish()
}
