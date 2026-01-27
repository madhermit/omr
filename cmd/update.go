package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const githubRepo = "madhermit/omr"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate checks GitHub for a newer version and prints a message if available
func CheckForUpdate() {
	if Version == "dev" {
		return // Skip check for dev builds
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
	if err != nil {
		return // Silently fail - don't interrupt user
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	if latest != current && latest > current {
		yellow := color.New(color.FgYellow).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Fprintf(os.Stderr, "\n%s Update available: %s → %s\n", yellow("!"), cyan(current), cyan(latest))

		// Detect installation method
		exe, _ := os.Executable()
		if strings.Contains(exe, "mise") {
			fmt.Fprintf(os.Stderr, "  Run: %s\n\n", cyan("mise upgrade omr"))
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n\n", cyan("https://github.com/"+githubRepo+"/releases/tag/v"+latest))
		}
	}
}
