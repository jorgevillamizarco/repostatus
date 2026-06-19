package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	dirFlag    = flag.String("dir", "", "directory to scan (default: ~/Documents/projects)")
	sigs       = make(chan os.Signal, 1)
)

type RepoInfo struct {
	Name   string
	Status string
	Branch string
}

func main() {
	flag.Parse()

	// Expand ~ in path
	dir := *dirFlag
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			home = ""
		}
	}
	if dir == "" {
		dir = filepath.Join(home, "Documents", "projects")
	}
	if strings.HasPrefix(dir, "~/") {
		dir = filepath.Join(home, dir[2:])
	}

	// Setup signal handling
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		os.Exit(0)
	}()

	// Scan directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	// Find git repos
	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		gitPath := filepath.Join(dir, entry.Name(), ".git")
		info, err := os.Stat(gitPath)
		if err != nil || !info.IsDir() {
			continue
		}
		repos = append(repos, entry.Name())
	}

	// Sort alphabetically
	sort.Strings(repos)

	// Get status for each repo
	var infos []RepoInfo
	for _, name := range repos {
		status, branch := getGitStatus(filepath.Join(dir, name))
		infos = append(infos, RepoInfo{
			Name:   name,
			Status: status,
			Branch: branch,
		})
	}

	// Print table
	printTable(infos)
}

func getGitStatus(repoDir string) (string, string) {
	// Check for bare repository
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--is-bare-repository")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "true" {
		return "bare", "—"
	}

	// Check for broken .git
	cmd = exec.Command("git", "-C", repoDir, "status", "--porcelain")
	statusOut, err := cmd.Output()
	if err != nil {
		return "error", "—"
	}

	// Check if empty (no commits)
	cmd = exec.Command("git", "-C", repoDir, "rev-list", "--count", "HEAD")
	out, err = cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) == "0" {
		return "empty", "—"
	}

	// Check detached HEAD
	cmd = exec.Command("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err = cmd.Output()
	branch := strings.TrimSpace(string(out))
	if err == nil && strings.TrimSpace(string(out)) == "HEAD" {
		cmd = exec.Command("git", "-C", repoDir, "rev-parse", "--short", "HEAD")
		out, err = cmd.Output()
		if err != nil {
			return "detached", "—"
		}
		short := strings.TrimSpace(string(out))
		// Check dirty state before returning
		porcelain := strings.TrimSpace(string(statusOut))
		dirtySuffix := ""
		if porcelain != "" {
			untracked := 0
			for _, line := range strings.Split(porcelain, "\n") {
				if len(line) >= 2 && line[0] == '?' {
					untracked++
				}
			}
			if untracked > 0 {
				dirtySuffix = "*"
			}
			return "detached" + dirtySuffix, short
		}
		return "detached", short
	}

	// Check for no remote
	cmd = exec.Command("git", "-C", repoDir, "remote")
	remoteOut, err := cmd.Output()
	if err != nil || len(strings.TrimSpace(string(remoteOut))) == 0 {
		return "no-remote", branch
	}

	// Check dirty
	porcelain := strings.TrimSpace(string(statusOut))

	// Check ahead/behind (before clean return — clean repos can be behind)
	cmd = exec.Command("git", "-C", repoDir, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	out, err = cmd.Output()
	aheadBehind := ""
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) == 2 {
			behind, _ := strconv.Atoi(parts[0])
			ahead, _ := strconv.Atoi(parts[1])
			if ahead > 0 && behind > 0 {
				aheadBehind = fmt.Sprintf(" diverged ↑%d↓%d", ahead, behind)
			} else if ahead > 0 {
				aheadBehind = fmt.Sprintf(" ahead ↑%d", ahead)
			} else if behind > 0 {
				aheadBehind = fmt.Sprintf(" behind ↓%d", behind)
			}
		}
	}

	if porcelain == "" {
		if aheadBehind != "" {
			return "clean" + aheadBehind, branch
		}
		return "clean", branch
	}

	// Count untracked files
	untracked := 0
	for _, line := range strings.Split(porcelain, "\n") {
		if len(line) >= 2 && line[0] == '?' {
			untracked++
		}
	}

	// Determine dirty status
	if untracked > 0 {
		return "dirty*" + aheadBehind, branch
	}
	return "dirty" + aheadBehind, branch
}

func colorize(status string) string {
	switch {
	case status == "clean":
		return "\033[32m"
	case strings.Contains(status, "dirty"):
		return "\033[33m"
	case strings.Contains(status, "behind"):
		return "\033[33m"
	case strings.Contains(status, "ahead"):
		return "\033[36m"
	case strings.Contains(status, "diverged"):
		return "\033[35m"
	case status == "no-remote" || status == "detached" || status == "empty":
		return "\033[31m"
	default:
		return "\033[31m"
	}
}

func printTable(infos []RepoInfo) {
	// Find longest name and branch for column widths
	nameW := 7    // len("PROJECT")
	statusW := 7  // len("STATUS")
	branchW := 6  // len("BRANCH")
	for _, info := range infos {
		n := utf8.RuneCountInString(info.Name)
		if n > nameW {
			nameW = n
		}
		s := utf8.RuneCountInString(info.Status)
		if s > statusW {
			statusW = s
		}
		b := utf8.RuneCountInString(info.Branch)
		if b > branchW {
			branchW = b
		}
	}
	nameW += 2
	statusW += 2
	branchW += 2

	reset := "\033[0m"
	pad := func(s string, w int) string {
		n := utf8.RuneCountInString(s)
		if n >= w {
			return s
		}
		return s + strings.Repeat(" ", w-n)
	}

	// Header
	fmt.Println(pad("PROJECT", nameW) + " " + pad("STATUS", statusW) + " " + pad("BRANCH", branchW))
	fmt.Println(strings.Repeat("─", nameW) + " " + strings.Repeat("─", statusW) + " " + strings.Repeat("─", branchW))

	// Rows
	for _, info := range infos {
		color := colorize(info.Status)
		statusField := color + pad(info.Status, statusW) + reset
		fmt.Println(pad(info.Name, nameW) + " " + statusField + " " + pad(info.Branch, branchW))
	}
}