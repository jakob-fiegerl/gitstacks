package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/magiconair/properties"
)

type ConfigMap struct {
	Username string
	Remote   string
}

func Execute(script string) string {
	command := exec.Command("bash")
	command.Stdin = strings.NewReader(script)
	out, err := command.Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	output := string(out[:])
	return strings.Replace(output, `\n`, "\n", -1)
}

func ExecuteCommand(args string) string {
	output, err := exec.Command("bash", "-c", args).Output()
	if err != nil {
		fmt.Printf("Unable to execute command {%s}. Error: %s\n", args, err)
	}
	return string(output)
}

func Open(url string) error {
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

func SetupConfig() *ConfigMap {
	configPath := getConfigPath()
	// Check if the config file exists
	if !doesFileExist(configPath) {
		// Create the config file
		createFile(configPath)
	}
	p := properties.MustLoadFile(getConfigPath(), properties.UTF8)
	return &ConfigMap{
		Remote:   GetRemote(),
		Username: p.MustGetString("username"),
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

func GetRemote() string {
	remote := ExecuteCommand("git config --get remote.origin.url")
	re := regexp.MustCompile(`git@github\.com:(.+)/(.+)\.git`)
	// Replace the SSH URL with the HTTPS URL using captured groups
	httpsURL := re.ReplaceAllString(remote, `https://github.com/$1/$2.git`)
	withoutGit := strings.Replace(httpsURL, ".git", "", 1)
	return TrimString(withoutGit)
}

func GetGitProvider() string {
	if strings.Contains(GetRemote(), "github") {
		return "github"
	}
	return "gitlab"
}
