package github

type commitCommentEvent struct {
	Comment    *github.RepositoryComment `json:"comment"`
	Repository *github.Repository        `json:"repository"`
}

func handleCommitCommentEvent(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event commitCommentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httpStatus(rw, http.StatusBadRequest)
		return
	}

	// A command is always placed at the beginning of the line
	// and it is prefixed with '!'.
	cmdRegexp := regexp.MustCompile("^[!]([a-zA-Z]+)(.*)$")

	// Process the comment body.
	scanner := bufio.NewScanner(strings.NewReader(*event.Comment.Body))
	for scanner.Scan() {
		// Check whether this is a command and continue if not.
		match := cmdRegexp.FindStringSubmatch(scanner.Text())
		if len(match) == 0 {
			continue
		}
		cmd, arg := match[1], strings.TrimSpace(match[2])

		var err error
		switch cmd {
		case "blocker":
			err = createReviewBlockerFromCommitComment(
				r,
				*event.Repository.Owner.Login,
				*event.Repository.Name,
				event.Comment,
				arg)
		}
		if err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		httputils.Error(rw, r, err)
		return
	}

	httpStatus(rw, http.StatusAccepted)
}

func createReviewBlockerFromCommitComment(
	r *http.Request,
	owner string,
	repo string,
	comment *github.RepositoryComment,
	blockerSummary string,
) error {

	// Get GitHub API client.
	client, err := githubutils.NewClient()
	if err != nil {
		return err
	}

	// Find the right review issue.
	//
	// We search the content of all review issues for the right commit hash.
	// This is not terribly robust but that is all we can do right now.
	//
	// GitHub shortens commit hashes to 7 leading characters, hence [:7].
	var (
		commitSHA     = *comment.CommitID
		commentURL    = *comment.HTMLURL
		commentAuthor = *comment.User.Login
		pattern       = fmt.Sprintf("] %v:", commitSHA[:7])
	)

	query := fmt.Sprintf(
		`"%v" repo:"%v/%v" type:issue state:open state:closed label:review in:body`,
		pattern, owner, repo)

	res, _, err := client.Search.Issues(query, &github.SearchOptions{})
	if err != nil {
		return err
	}
	if num := *res.Total; num != 1 {
		log.Warn(r, "failed to find the review issue for commit %v (%v issues found)",
			commitSHA, num)
		return nil
	}
	issue := res.Issues[0]

	// Parse issue body.
	issueCtx, err := ghissues.ParseReviewIssue(&issue)
	if err != nil {
		return err
	}

	// Add the new review issue record.
	issueCtx.AddReviewBlocker(commitSHA, commentURL, blockerSummary, false)

	// Update the review issue.
	issueNum := *issue.Number
	_, _, err = client.Issues.Edit(owner, repo, issueNum, &github.IssueRequest{
		Body:  github.String(issueCtx.FormatBody()),
		State: github.String("open"),
	})
	if err != nil {
		return err
	}

	log.Info(r, "Linked a new review comment to review issue %v/%v#%v", owner, repo, issueNum)

	// Add the blocker comment.
	body := fmt.Sprintf("A new [review blocker](%v) was opened by @%v for review issue #%v.",
		commentURL, commentAuthor, issueNum)

	_, _, err = client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: github.String(body),
	})
	return err
}
