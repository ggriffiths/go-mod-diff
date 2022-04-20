package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/colorstring"
	"github.com/radeksimko/go-mod-diff/diff"
	"github.com/radeksimko/go-mod-diff/github"
	"github.com/radeksimko/go-mod-diff/gomod"
	"github.com/radeksimko/go-mod-diff/govendor"
)

func main() {
	// Setup GitHub connection
	gh := github.NewGitHub()
	if os.Getenv("GITHUB_TOKEN") != "" {
		gh = github.NewGitHubWithToken(os.Getenv("GITHUB_TOKEN"))
	}

	// Parse go modules file
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	goModFile, err := gomod.ParseFile(filepath.Join(cwd, "go.mod"))

	// Parse govendor file
	govendorFile, err := govendor.ParseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Compare both and print out differences
	d, err := diff.CompareGoModWithGovendor(goModFile, govendorFile, gh)
	if err != nil {
		log.Fatal(err)
	}

	printDifference(d, gomod.GetVersionForModule(goModFile))

	total := len(goModFile.Require) - len(d.Matched)

	colorstring.Printf("\n\nMatched package revisions: [bold][green]%d[reset] of %d.\n"+
		"[bold]%d[reset] to check ([bold][red]%d[reset] not found and [bold][yellow]%d[reset] different revs).\n",
		len(d.Matched), len(goModFile.Require), total, len(d.NotFound), len(d.Different))
}

func printDifference(d *diff.Diff, vlF gomod.VersionLookupFunc) {
	for _, entry := range d.Errored {
		printDiffEntry(entry, vlF)
	}

	for _, entry := range d.NotFound {
		printDiffEntry(entry, vlF)
	}

	for _, entry := range d.Different {
		printDiffEntry(entry, vlF)
	}

	for _, entry := range d.Matched {
		colorstring.Printf("\n[bold]%s[reset] [bold][green]✓[reset]", entry.ModulePath)
	}
}

func printDiffEntry(de *diff.DiffEntry, vlF gomod.VersionLookupFunc) {
	mts, stderr, err := gomod.GoModWhy(de.ModulePath)
	if err != nil {
		colorstring.Printf("[bold][red]Failed to check (%s)[reset][red]\n%s", err, stderr)
		return
	}

	var govendorVerStr string
	if len(de.GoVendorVersions) == 0 {
		govendorVerStr = "notFound"
	} else {
		govendorVerStr = de.GoVendorVersions[0].String()
	}
	colorstring.Printf("%s,%s,%s,%v\n", de.ModulePath, de.GoModVersion.String(), govendorVerStr, mts)

}
