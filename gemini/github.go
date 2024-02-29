package gemini

import (
	"context"
	"fmt"
	"mime"
	"path"
	"strconv"
	"strings"

	// "net/url"

	"github.com/google/go-github/v35/github"
	"gitlab.com/clseibold/auragem_sis/config"
	sis "gitlab.com/clseibold/smallnetinformationservices"
	"golang.org/x/oauth2"
)

var apiToken = config.GithubToken

func handleGithub(g sis.ServerHandle) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	g.AddRoute("/github", func(request sis.Request) {
		request.Gemini(`# AuraGem Github Proxy

Welcome to the AuraGem Github proxy!

=> /github/search Search Repos
`)
	})

	g.AddRoute("/github/search/:page", func(request sis.Request) {
		//query, err2 := c.QueryString()
		query := request.Query()
		/*if err2 != nil {
			return err2
		} else*/if query == "" {
			request.RequestInput("Search Query:")
			//return c.NoContent(gig.StatusInput, "Search Query:")
		} else {
			handleGithubSearch(ctx, request, client, query, request.GetParam("page"))
		}
	})
	g.AddRoute("/github/search", func(request sis.Request) {
		//query, err2 := c.QueryString()
		query := request.Query()
		/*if err2 != nil {
			return err2
		} else*/if query == "" {
			request.RequestInput("Search Query:")
			//return c.NoContent(gig.StatusInput, "Search Query:")
		} else {
			handleGithubSearch(ctx, request, client, query, "")
		}
	})

	g.AddRoute("/github/repo/:id", func(request sis.Request) {
		id := request.GetParam("id")
		template := `# Repo: %s

%s

SSH Url: %s
HTML Url: %s
Homepage: %s
=> %s License: %s

=> /github/repo/%d/issues/ Issues
=> /github/repo/%d/b Branches

## Contents - Branch %s

%s
`
		id_int, err1 := strconv.Atoi(id)
		if err1 != nil {
			panic(err1)
		}

		repository, _, err2 := client.Repositories.GetByID(ctx, int64(id_int))
		if err2 != nil {
			panic(err2)
		}

		// TODO: README, README.md, readme.md
		/*opts := &github.RepositoryContentGetOptions{}
		readmeContents, _, _, err3 := client.Repositories.GetContents(ctx, repository.GetOwner().GetLogin(), repository.GetName(), "README.md", opts)
		readmeContents_str := ""
		var err4 error
		if err3 != nil {
			readmeContents, _, _, err3 = client.Repositories.GetContents(ctx, *repository.GetOwner().Login, repository.GetName(), "readme.md", opts)
			if err3 == nil {
				readmeContents_str, err4 = readmeContents.GetContent()
			}
		} else {
			readmeContents_str, err4 = readmeContents.GetContent()
		}
		if err4 != nil {
			panic(err4)
		}*/

		rootContents, _ := getRepoContents(ctx, client, repository, "")
		request.Gemini(fmt.Sprintf(template, repository.GetFullName(), repository.GetDescription(), repository.GetSSHURL(), repository.GetHTMLURL(), repository.GetHomepage(), repository.GetLicense().GetURL(), repository.GetLicense().GetName(), repository.GetID(), repository.GetID(), repository.GetDefaultBranch(), rootContents))
	})

	g.AddRoute("/github/repo/:id/b", func(request sis.Request) {
		id := request.GetParam("id")

		id_int, err1 := strconv.Atoi(id)
		if err1 != nil {
			panic(err1)
		}

		repository, _, err2 := client.Repositories.GetByID(ctx, int64(id_int))
		if err2 != nil {
			panic(err2)
		}

		template := `# Repo: %s - Branches

=> /github/repo/%d Repo Home

%s`

		opts2 := &github.BranchListOptions{}
		branches, _, err3 := client.Repositories.ListBranches(ctx, repository.GetOwner().GetLogin(), repository.GetName(), opts2)
		if err3 != nil {
			panic(err3)
		}

		var builder strings.Builder
		for _, branch := range branches {
			fmt.Fprintf(&builder, "=> /github/repo/%d/b/%s %s\n", repository.GetID(), branch.GetName(), branch.GetName())
		}

		request.Gemini(fmt.Sprintf(template, repository.GetFullName(), repository.GetID(), builder.String()))
	})

	g.AddRoute("/github/repo/:id/issues/", func(request sis.Request) {
		id := request.GetParam("id")

		id_int, err1 := strconv.Atoi(id)
		if err1 != nil {
			panic(err1)
		}

		repository, _, err2 := client.Repositories.GetByID(ctx, int64(id_int))
		if err2 != nil {
			panic(err2)
		}

		//client.Repositories.ListCommits()
		//client.Repositories.ListReleases()

		// Opts for filtering: Milestone (number, none, *), State (open, closed, all), Sort (updated, created, comments)
		opts := &github.IssueListByRepoOptions{State: "open", Sort: "updated", Direction: "desc"}
		opts.PerPage = 100
		issues, _, err := client.Issues.ListByRepo(ctx, repository.GetOwner().GetLogin(), repository.GetName(), opts)
		if err != nil {
			panic(err) // TODO
		}

		var builder strings.Builder
		for _, issue := range issues {
			fmt.Fprintf(&builder, "=> /github/repo/%d/issues/%d %s #%d: %s (%s)\n", repository.GetID(), issue.GetNumber(), issue.GetCreatedAt().Format("2006-01-02"), issue.GetNumber(), issue.GetTitle(), issue.User.GetLogin())
		}

		request.Gemini(fmt.Sprintf(`# %s Open Issues (%d)

=> /github/repo/%d Repo Home

%s
`, repository.GetFullName(), len(issues), repository.GetID(), builder.String()))
	})

	g.AddRoute("/github/repo/:id/issues/:issue", func(request sis.Request) {
		id := request.GetParam("id")
		issueParam := request.GetParam("issue")

		id_int, err1 := strconv.Atoi(id)
		if err1 != nil {
			panic(err1)
		}
		issue_int, err2 := strconv.Atoi(issueParam)
		if err2 != nil {
			panic(err2)
		}

		repository, _, err2 := client.Repositories.GetByID(ctx, int64(id_int))
		if err2 != nil {
			panic(err2)
		}

		issue, _, err := client.Issues.Get(ctx, repository.GetOwner().GetLogin(), repository.GetName(), issue_int)
		if err != nil {
			panic(err) // TODO
		}

		// Get Comments on the Issue
		sort := "created"
		direction := "asc"
		opts := &github.IssueListCommentsOptions{Sort: &sort, Direction: &direction}
		comments, _, err3 := client.Issues.ListComments(ctx, repository.GetOwner().GetLogin(), repository.GetName(), issue.GetNumber(), opts)
		if err3 != nil {
			panic(err3)
		}

		var builder strings.Builder
		for _, comment := range comments {
			fmt.Fprintf(&builder, "### %s %s\n\n%s\n\n", comment.GetCreatedAt().Format("2006-01-02 15:04:05"), comment.GetUser().GetLogin(), comment.GetBody())
		}

		request.Gemini(fmt.Sprintf(`# %s Issue #%d: %s

=> /github/repo/%d Repo Home
=> /github/repo/%d/issues/ Issues

## %s %s

%s

## Comments (%d)

%s
`, repository.GetFullName(), issue.GetNumber(), issue.GetTitle(), repository.GetID(), repository.GetID(), issue.GetCreatedAt().Format("2006-01-02 15:04:05"), issue.GetUser().GetLogin(), issue.GetBody(), len(comments), builder.String()))
	})

	g.AddRoute("/github/repo/:id/files/*", func(request sis.Request) {
		id := request.GetParam("id")
		route := fmt.Sprintf("/github/repo/%s/files", id)
		p := strings.Replace(request.RawPath(), route, "", 1)

		id_int, err1 := strconv.Atoi(id)
		if err1 != nil {
			panic(err1)
		}

		repository, _, err2 := client.Repositories.GetByID(ctx, int64(id_int))
		if err2 != nil {
			panic(err2)
		}

		template := `# Repo Contents: %s - Branch %s

Path: %s

..
%s
`
		contents, isFile := getRepoContents(ctx, client, repository, p)
		if isFile {
			if strings.HasSuffix(p, ".gmi") || strings.HasSuffix(p, ".gemini") {
				request.Gemini(contents)
			} else if strings.HasSuffix(p, ".md") {
				request.TextWithMimetype("text/markdown", contents)
			} else if strings.HasSuffix(p, ".rss") || strings.HasSuffix(p, ".atom") {
				request.Bytes("text/rss", []byte(contents))
			} else if strings.HasSuffix(p, ".gpub") {
				request.Bytes("application/gpub+zip", []byte(contents))
			} else {
				extension := path.Ext(p)
				mimeType := mime.TypeByExtension(extension)
				if mimeType == "" {
					//return c.Text(contents)
					request.TextWithMimetype("text/plain", contents)
				} else {
					request.TextWithMimetype(mimeType, contents)
				}
			}
		} else {
			request.Gemini(fmt.Sprintf(template, repository.GetFullName(), repository.GetDefaultBranch(), p, contents))
		}
	})
}

func getRepoContents(ctx context.Context, client *github.Client, repository *github.Repository, path string) (string, bool) {
	opts := &github.RepositoryContentGetOptions{}
	fileContents, dirContents, _, err := client.Repositories.GetContents(ctx, repository.GetOwner().GetLogin(), repository.GetName(), path, opts)
	if err != nil {
		//panic(err) // TODO
		return "Not found.", false
	}

	var builder strings.Builder
	if dirContents != nil {
		for _, v := range dirContents {
			// TODO: Use PathEscape for each part of the path, but cannot escape the whole path (otherwise "/" will turn into "%2F")
			// OR, switch this to use a query string so that I can use QueryEscape?
			fmt.Fprintf(&builder, "=> /github/repo/%d/files/%s %s\n", repository.GetID(), v.GetPath(), v.GetName())
		}
		return builder.String(), false
	}
	if fileContents != nil {
		c, _ := fileContents.GetContent()
		return c, true
	}
	if dirContents == nil && fileContents == nil {
		fmt.Fprintf(&builder, "Not found")
		return builder.String(), false
	}

	return builder.String(), false
}

func handleGithubSearch(ctx context.Context, request sis.Request, client *github.Client, query string, page string) {
	template := `# Github Search

=> /github/search New Search

%s`

	opts := &github.SearchOptions{}
	result, _, err := client.Search.Repositories(ctx, query, opts)
	if err != nil {
		panic(err)
	}

	var builder strings.Builder
	for _, repository := range result.Repositories {
		fmt.Fprintf(&builder, "=> /github/repo/%d %s\n%s\n\n", repository.GetID(), repository.GetFullName(), repository.GetDescription())
	}

	request.Gemini(fmt.Sprintf(template, builder.String()))
}
