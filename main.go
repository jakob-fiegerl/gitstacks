package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"

	"fiegerl.at/gitstacks/internal"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	info(app)
	commands(app)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func info(app *cli.App) {
	app.Name = "GitStacks"
	app.Usage = "An opinionated cli tool for advanced git commands"
	app.Authors = []*cli.Author{
		{
			Name: "Jakob Fiegerl",
		},
	}
	app.Version = "1.0.0"
}

func commands(app *cli.App) {
	config := internal.SetupConfig()
	remote := config.Remote
	username := config.Username

	app.Commands = []*cli.Command{
		{
			Name:  "mr",
			Usage: "open assigned merge requests",
			Action: func(c *cli.Context) error {
				provider := internal.GetGitProvider()
				if provider == "gitlab" {
					internal.Open(remote + "/dashboard/merge_requests?assignee_username=" + username)
				} else if provider == "github" {
					internal.Open(remote + "/pulls/" + username)
				}
				return nil
			},
		},
		{
			Name:  "remote",
			Usage: "opens remote repository in browser",
			Action: func(c *cli.Context) error {
				internal.Open(remote)
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:  "mr",
					Usage: "opens remote repository mrs",
					Action: func(c *cli.Context) error {
						internal.Open(remote + "/-/merge_requests")
						return nil
					},
				},
			},
		},
		{
			Name:  "tags",
			Usage: "list the 5 latest tags",
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "list the 5 latest tags",
					Action: func(c *cli.Context) error {
						fmt.Println(internal.ExecuteCommand("git for-each-ref --sort='*authordate' --format='%(tag) | %(taggerdate:short)' refs/tags | tail -5"))
						return nil
					},
				},
				{
					Name:  "create",
					Usage: "creates a tag with the latest unreleased commits",
					Action: func(c *cli.Context) error {
						unreleased := "" //execute(getUnreleasedCommits)
						tag := c.Args().First()
						internal.ExecuteCommand("git tag -a " + tag + " -m \"" + unreleased + "\"")
						internal.ExecuteCommand("git push origin " + tag)
						fmt.Println("Tag created: " + tag)
						fmt.Println("Link: " + remote + "/-/tags/" + tag)
						return nil
					},
				},
				{
					Name:  "diff",
					Usage: "show the difference between two tags",
					Action: func(c *cli.Context) error {
						tag1 := strings.TrimSuffix(internal.Execute("git describe --tags $(git rev-list --tags --max-count=1)"), "\n")
						tag2 := "main"
						fmt.Println(internal.ExecuteCommand("git log " + tag1 + ".." + tag2 + " --no-merges --pretty=format:\"%h - %an, %ad : %s\" --date=short"))
						return nil
					},
				},
			},
		},
		{
			Name:    "save",
			Usage:   "saves the current changes",
			Aliases: []string{"s"},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "push",
					Value:   false,
					Aliases: []string{"p"},
				},
			},
			Action: func(c *cli.Context) error {
				message := c.Args().First()
				push := c.Bool("push")
				if message == "" {
					message = "work in progress"
				} else {
					message = strings.ReplaceAll(message, "\"", "")
				}
				pushCommand := ""
				if push {
					pushCommand = " && git push"
				}
				fmt.Println(internal.ExecuteCommand("git add --all && git commit -a -m \"" + message + "\"" + pushCommand))
				fmt.Println("[GitStacks] Saved changes")
				return nil
			},
		},
		{
			Name:  "new",
			Usage: "creates a new stack from a given branch (default main)",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "branch",
					Value:   "main",
					Aliases: []string{"b"},
				},
			},
			Action: func(c *cli.Context) error {
				branch := c.String("branch")
				baseCommand := "git add --all && git stash && git checkout " + branch + " && git pull && git checkout -b " + c.Args().First()
				changedFiles := internal.TrimWhitespaceAndNewline(internal.ExecuteCommand("git diff --name-only | wc -l"))
				if changedFiles == "0" {
					fmt.Println("[GitStacks] No changes to save")
				} else {
					baseCommand += " && git stash pop"
				}
				fmt.Println(internal.ExecuteCommand(baseCommand))
				return nil
			},
		},
	}
}
