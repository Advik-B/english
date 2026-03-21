package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Advik-B/english/version"
	"github.com/spf13/cobra"
)

const githubReleasesURL = "https://api.github.com/repos/Advik-B/english/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

// parseVersion splits a "major.minor.patch" string into its three integer
// components. The optional leading "v" is stripped before parsing.
func parseVersion(v string) (major, minor, patch int, err error) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version %q", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	patch, err = strconv.Atoi(parts[2])
	return
}

// isNewer reports whether candidate is strictly newer than base using
// semantic-style integer comparison (major → minor → patch).
func isNewer(base, candidate string) bool {
	bMaj, bMin, bPat, err1 := parseVersion(base)
	cMaj, cMin, cPat, err2 := parseVersion(candidate)
	if err1 != nil || err2 != nil {
		// Fall back to plain string comparison if parsing fails.
		return candidate != base
	}
	if cMaj != bMaj {
		return cMaj > bMaj
	}
	if cMin != bMin {
		return cMin > bMin
	}
	return cPat > bPat
}

// CheckForUpdates queries the GitHub releases API and prints whether a newer
// version of the English interpreter is available. It never downloads or
// installs anything automatically.
func CheckForUpdates() {
	fmt.Print("Checking for updates... ")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(githubReleasesURL)
	if err != nil {
		fmt.Println("failed.")
		fmt.Println("Could not reach GitHub:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("failed.")
		fmt.Printf("GitHub returned status %d\n", resp.StatusCode)
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Println("failed.")
		fmt.Println("Could not parse response:", err)
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := version.Version

	if isNewer(current, latest) {
		fmt.Println("update available!")
		fmt.Printf("Current version : %s\n", current)
		fmt.Printf("Latest version  : %s\n", latest)
		fmt.Printf("Download        : %s\n", release.HTMLURL)
	} else {
		fmt.Println("up to date.")
		fmt.Printf("You are running the latest version (%s).\n", current)
	}
}

var updatesCmd = &cobra.Command{
	Use:     "updates",
	Aliases: []string{"check-for-updates"},
	Short:   "Check for a newer version of the English interpreter",
	Long: `Queries the GitHub releases page and reports whether a newer version is
available. This command never downloads or installs anything automatically.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckForUpdates()
	},
}
