package generator

// This file contains template and generation helpers that are specific to
// GitHub-hosted repositories and GitHub-flavoured outputs.
//
// Where relevant, the helpers below follow GitHub's documented repository and
// feature URL shapes rather than treating the repository host as a generic
// forge.

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var gitHubFuncMap = template.FuncMap{
	"isHostedOnGitHub":                             isHostedOnGitHub,
	"repositoryLocationOnGitHub":                   repositoryLocationOnGitHub,
	"repositoryImportPathOnGitHub":                 repositoryImportPathOnGitHub,
	"githubPagesURL":                               githubPagesURL,
	"projectOwnGitHubActionsLocation":              projectOwnGitHubActionsLocation,
	"githubIssuesURL":                              githubIssuesURL,
	"githubPullsURL":                               githubPullsURL,
	"githubDiscussionsURL":                         githubDiscussionsURL,
	"githubAuthenticationForGoPrivateDependencies": githubAuthenticationForGoPrivateDependencies,
}

const (
	defaultGitHubOrganisation = "github.com/Arm-Debug"
)

// parseGitHubRepositoryURL accepts the canonical HTTPS repository URL shape and
// extracts the `<organisation>/<repoName>` location used by other GitHub-specific helpers.
// Repository URL formats are described in GitHub's cloning documentation:
// https://docs.github.com/repositories/creating-and-managing-repositories/cloning-a-repository
func parseGitHubRepositoryURL(raw string) (repositoryLocation, organisation, repoName string, isGitHub bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil {
		return
	}
	if !strings.EqualFold(parsed.Host, "github.com") {
		return
	}

	trimmedPath := strings.Trim(parsed.Path, urlPathSeparator)
	if reflection.IsEmpty(trimmedPath) {
		return
	}

	parts := strings.Split(trimmedPath, urlPathSeparator)
	if len(parts) != 2 {
		return
	}

	organisation = strings.TrimSpace(parts[0])
	repoName = strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
	if collection.AnyTrue(reflection.IsEmpty(organisation), reflection.IsEmpty(repoName)) {
		return
	}

	repositoryLocation = path.Join(organisation, repoName)
	isGitHub = true
	return
}

func isHostedOnGitHub(spec *SDKSpecification) bool {
	_, _, _, ok := parseGitHubRepositoryURL(repositoryURL(spec))
	return ok
}

func repositoryLocationOnGitHub(spec *SDKSpecification) string {
	repo, _, _, _ := parseGitHubRepositoryURL(repositoryURL(spec))
	return repo
}

func repositoryImportPathOnGitHub(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("github.com/%s", repo)
	}
	return ""
}

// projectOwnGitHubActionsLocation returns the repository-qualified `actions`
// path used in GitHub Actions `uses:` references for actions published from the
// project itself.
func projectOwnGitHubActionsLocation(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("%s/actions", repo)
	}
	return ""
}

func repositoryOrganisationOnGitHub(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		parts := strings.Split(repo, urlPathSeparator)
		if len(parts) == 2 {
			return parts[0]
		}
	}
	return ""
}

func repositoryNameOnGitHub(spec *SDKSpecification) string {
	_, _, repoName, _ := parseGitHubRepositoryURL(repositoryURL(spec))
	return repoName
}

func githubPagesURL(spec *SDKSpecification) string {
	organisation := strings.ToLower(repositoryOrganisationOnGitHub(spec))
	repoName := repositoryNameOnGitHub(spec)
	if collection.AnyTrue(reflection.IsEmpty(organisation), reflection.IsEmpty(repoName)) {
		return ""
	}
	return fmt.Sprintf("https://%s.github.io/%s", organisation, repoName)
}

// githubIssuesURL returns the standard GitHub Issues URL for the repository.
// GitHub Issues are documented at:
// https://docs.github.com/en/issues/tracking-your-work-with-issues/about-issues
func githubIssuesURL(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("https://github.com/%s/issues", repo)
	}
	return ""
}

// githubPullsURL returns the standard GitHub Pull Requests URL for the
// repository. Pull request concepts are documented at:
// https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-with-pull-requests/about-pull-requests
func githubPullsURL(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("https://github.com/%s/pulls", repo)
	}
	return ""
}

// githubDiscussionsURL returns the standard GitHub Discussions URL for the
// repository. Discussions are documented at:
// https://docs.github.com/en/discussions/collaborating-with-your-community-using-discussions/about-discussions
func githubDiscussionsURL(spec *SDKSpecification) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("https://github.com/%s/discussions", repo)
	}
	return ""
}

// the linkToFileOnGitHub returns the URL used by GitHub's file browser.
// See https://docs.github.com/en/repositories/working-with-files/using-files/getting-permanent-links-to-files
func linkToFileOnGitHub(spec *SDKSpecification, relativePath string) string {
	if repo := repositoryLocationOnGitHub(spec); !reflection.IsEmpty(repo) {
		return fmt.Sprintf("https://github.com/%s/blob/main/%s", repo, strings.TrimPrefix(relativePath, urlPathSeparator))
	}
	return ""
}

// githubAuthenticationForGoPrivateDependencies returns the Git configuration
// command used to inject an access token for private Go dependencies that
// are expected to live under the same GitHub organisation as the current project.
func githubAuthenticationForGoPrivateDependencies(spec *SDKSpecification) string {
	organisation := repositoryOrganisationOnGitHub(spec)
	if reflection.IsEmpty(organisation) {
		return `echo "nothing to do"`
	}

	return fmt.Sprintf("git config --global url.\"https://${{ secrets.GIT_SECRET }}:x-oauth-basic@github.com/%s\".insteadOf \"https://github.com/%s\"", organisation, organisation)
}
