package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/magiconair/properties"
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

type ConfigMap struct {
	username string
	remote   string
}

func commands(app *cli.App) {
	config := setupConfig()
	remote := config.remote
	username := config.username

	app.Commands = []*cli.Command{
		{
			Name:  "mr",
			Usage: "open assigned merge requests",
			Action: func(c *cli.Context) error {
				provider := getGitProvider()
				if provider == "gitlab" {
					open(remote + "/dashboard/merge_requests?assignee_username=" + username)
				} else if provider == "github" {
					open(remote + "/pulls/" + username)
				}
				return nil
			},
		},
		{
			Name:  "remote",
			Usage: "opens remote repository in browser",
			Action: func(c *cli.Context) error {
				open(remote)
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:  "mr",
					Usage: "opens remote repository mrs",
					Action: func(c *cli.Context) error {
						open(remote + "/-/merge_requests")
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
						fmt.Println(executeCommand("git for-each-ref --sort='*authordate' --format='%(tag) | %(taggerdate:short)' refs/tags | tail -5"))
						return nil
					},
				},
				{
					Name:  "create",
					Usage: "creates a tag with the latest unreleased commits",
					Action: func(c *cli.Context) error {
						unreleased := "" //execute(getUnreleasedCommits)
						tag := c.Args().First()
						executeCommand("git tag -a " + tag + " -m \"" + unreleased + "\"")
						executeCommand("git push origin " + tag)
						fmt.Println("Tag created: " + tag)
						fmt.Println("Link: " + remote + "/-/tags/" + tag)
						return nil
					},
				},
				{
					Name:  "diff",
					Usage: "show the difference between two tags",
					Action: func(c *cli.Context) error {
						tag1 := strings.TrimSuffix(execute("git describe --tags $(git rev-list --tags --max-count=1)"), "\n")
						tag2 := "main"
						fmt.Println(executeCommand("git log " + tag1 + ".." + tag2 + " --no-merges --pretty=format:\"%h - %an, %ad : %s\" --date=short"))
						return nil
					},
				},
			},
		},
		{
			Name:    "git",
			Aliases: []string{"g"},
			Usage:   "all git commands",
			Subcommands: []*cli.Command{
				{
					Name:  "switch",
					Usage: "switch to a branch",
					Action: func(c *cli.Context) error {
						fmt.Println(executeCommand("git add --all && git stash && git checkout -b " + c.Args().First() + " && git stash pop"))
						return nil
					},
				},
				{
					Name:  "branch",
					Usage: "get current branch",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "delete, d",
							Usage: "Delte the current branch",
						},
					},
					Action: func(c *cli.Context) error {
						branch := executeCommand("git rev-parse --abbrev-ref HEAD")
						if !c.Bool("delete") {
							fmt.Println(branch)
							return nil
						}
						if branch == "main" {
							// fmt.Println("[GitStacks] Cannot delete main branch")
							return cli.NewExitError("[GitStacks] Cannot delete main branch", 86)
						}
						fmt.Print("[GitStacks] Discard local changes (y|n): ")
						input := bufio.NewScanner(os.Stdin)
						input.Scan()
						if input.Text() == "y" {
							executeCommand("git reset --hard")
							executeCommand("git checkout main")
							executeCommand("git branch -D " + branch)
							fmt.Println("[GitStacks] Deleted branch: " + branch)
						}
						return nil
					},
				},
				{
					Name:  "save",
					Usage: "saves the current changes",
					Action: func(c *cli.Context) error {
						fmt.Println(executeCommand("git add --all && git commit -a -m \"work in progress\""))
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
				fmt.Println(executeCommand("git add --all && git commit -a -m \"" + message + "\"" + pushCommand))
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
				fmt.Println(executeCommand("git add --all && git stash && git checkout " + branch + " && git pull && git checkout -b " + c.Args().First() + " && git stash pop"))
				return nil
			},
		},
	}
}

func execute(script string) string {
	command := exec.Command("bash")
	command.Stdin = strings.NewReader(script)
	out, err := command.Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	output := string(out[:])
	return strings.Replace(output, `\n`, "\n", -1)
}

func executeCommand(args string) string {
	output, err := exec.Command("bash", "-c", args).Output()
	if err != nil {
		fmt.Printf("Unable to execute command {%s}. Error: %s\n", args, err)
	}
	return string(output)
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func TrimString(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

func setupConfig() *ConfigMap {
	configPath := getConfigPath()
	// Check if the config file exists
	if !doesFileExist(configPath) {
		// Create the config file
		createFile(configPath)
	}
	p := properties.MustLoadFile(getConfigPath(), properties.UTF8)
	return &ConfigMap{
		remote:   getRemote(),
		username: p.MustGetString("username"),
	}
}
func getConfigPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath + "/.gitstacks.properties"
}
func doesFileExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
func createFile(path string) {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func getRemote() string {
	remote := executeCommand("git config --get remote.origin.url")
	re := regexp.MustCompile(`git@github\.com:(.+)/(.+)\.git`)
	// Replace the SSH URL with the HTTPS URL using captured groups
	httpsURL := re.ReplaceAllString(remote, `https://github.com/$1/$2.git`)
	withoutGit := strings.Replace(httpsURL, ".git", "", 1)
	return TrimString(withoutGit)
}

func getGitProvider() string {
	if strings.Contains(getRemote(), "github") {
		return "github"
	}
	return "gitlab"
}
