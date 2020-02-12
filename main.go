package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	clilog "github.com/b4b4r07/go-cli-log"
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// These variables are set in build step
var (
	Version  = "unset"
	Revision = "unset"
)

// CLI represents this application itself
type CLI struct {
	Option Option
	Stdout io.Writer
	Stderr io.Writer
}

// Option represents application options
type Option struct {
	Number     int    `long:"number" description:"Number"`
	Repository string `long:"repository" description:"user/repo"`
	Body       string `long:"body" description:"Body"`
	Version    bool   `long:"version" description:"Show version"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	clilog.Env = "CLI_LOG"
	clilog.SetOutput()
	defer log.Printf("[INFO] finish main function")

	log.Printf("[INFO] Version: %s (%s)", Version, Revision)
	log.Printf("[INFO] Args: %#v", args)

	var opt Option
	args, err := flags.ParseArgs(&opt, args)
	if err != nil {
		return 2
	}

	cli := CLI{
		Option: opt,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := cli.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return 0
}

func (c *CLI) Run(args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		return errors.New("GITHUB_TOKEN is missing")
	}

	// Construct github HTTP client
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	slugs := strings.Split(c.Option.Repository, "/")
	if len(slugs) != 2 {
		return fmt.Errorf("repository %s should be like this style: user/repo", c.Option.Repository)
	}
	owner := slugs[0]
	repo := slugs[1]
	number := c.Option.Number
	body := c.Option.Body

	// Check there are no same comments
	comments, _, err := client.Issues.ListComments(context.Background(), owner, repo, number, nil)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		if comment.GetBody() == body {
			log.Printf("[INFO] comment %q was already posted, skip it", body)
			continue
		}
	}

	if _, _, err := client.Issues.CreateComment(context.Background(), owner, repo, number, &github.IssueComment{
		Body: &body,
	}); err != nil {
		return err
	}

	log.Printf("[INFO] Successfully created a comment!")
	return nil
}
