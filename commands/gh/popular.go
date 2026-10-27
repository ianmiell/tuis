package main

import "strings"

// Popular marks the commands people reach for most, each with a typical
// invocation. Curated from the gh manual and common cheat sheets;
// `go test` checks every entry still exists in the installed gh.
var Popular = map[string]string{
	"auth login":     "gh auth login",
	"auth status":    "gh auth status",
	"repo clone":     "gh repo clone cli/cli",
	"repo create":    "gh repo create my-project --public --clone",
	"repo view":      "gh repo view cli/cli --web",
	"repo fork":      "gh repo fork cli/cli --clone",
	"repo list":      "gh repo list --limit 10",
	"issue list":     "gh issue list --label bug --state open",
	"issue create":   "gh issue create --title \"Bug\" --body \"Details\"",
	"issue view":     "gh issue view 123 --comments",
	"issue close":    "gh issue close 123",
	"issue comment":  "gh issue comment 123 --body \"Thanks\"",
	"pr list":        "gh pr list --author \"@me\"",
	"pr create":      "gh pr create --fill",
	"pr view":        "gh pr view 123 --web",
	"pr checkout":    "gh pr checkout 123",
	"pr checks":      "gh pr checks 123",
	"pr diff":        "gh pr diff 123",
	"pr merge":       "gh pr merge 123 --squash --delete-branch",
	"pr status":      "gh pr status",
	"run list":       "gh run list --limit 5",
	"run view":       "gh run view --log-failed",
	"run watch":      "gh run watch",
	"workflow run":   "gh workflow run ci.yml",
	"release list":   "gh release list",
	"release create": "gh release create v1.0.0 --generate-notes",
	"gist create":    "gh gist create notes.txt --public",
	"api":            "gh api repos/cli/cli/issues --jq '.[].title'",
	"search issues":  "gh search issues --owner cli --state open",
	"search prs":     "gh search prs --author \"@me\" --state open",
	"status":         "gh status",
	"browse":         "gh browse",
	"alias set":      "gh alias set co 'pr checkout'",
	"config set":     "gh config set editor vim",
}

// PopularFlags are the flags most worth knowing, by long name, anywhere.
var PopularFlags = map[string]bool{
	"--repo": true, "--state": true, "--label": true, "--assignee": true,
	"--author": true, "--limit": true, "--json": true, "--jq": true,
	"--title": true, "--body": true, "--web": true, "--fill": true,
	"--draft": true, "--base": true, "--head": true, "--search": true,
}

func isPopular(path []string) bool {
	_, ok := Popular[strings.Join(path, " ")]
	return ok
}

// hasPopularBelow reports whether any popular command lives under path.
func hasPopularBelow(path []string) bool {
	p := strings.Join(path, " ") + " "
	for k := range Popular {
		if strings.HasPrefix(k+" ", p) || k == strings.TrimSpace(p) {
			return true
		}
	}
	return false
}
