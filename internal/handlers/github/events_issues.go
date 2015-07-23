package github

func handleIssuesEvent(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event github.IssueActivityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httpStatus(rw, http.StatusBadRequest)
		return
	}
	issue := event.Issue

	// Make sure this is a review issue event.
	var isReviewIssue bool
	for _, label := range issue.Labels {
		if *label.Name == "review" {
			isReviewIssue = true
			break
		}
	}
	if !isReviewIssue {
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Do nothing unless this is an opened, closed or reopened event.
	switch *event.Action {
	case "opened":
	case "closed":
	case "reopened":
	default:
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Parse issue body.
	issueCtx, err := ghissues.ParseReviewIssue(issue)
	if err != nil {
		log.Error(r, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// We are done in case this is a commit review issue.
	ctx, ok := issueCtx.(*ghissues.StoryReviewIssue)
	if !ok {
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Instantiate the issue tracker.
	tracker, err := trackers.GetIssueTracker(ctx.TrackerName)
	if err != nil {
		log.Error(r, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// Find relevant story.
	story, err := tracker.FindStoryByTag(ctx.StoryKey)
	if err != nil {
		log.Error(r, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// Invoke relevant event handler.
	var (
		issueNum = strconv.Itoa(*issue.Number)
		issueURL = *issue.HTMLURL
		ex       error
	)
	switch *event.Action {
	case "opened":
		ex = story.OnReviewRequestOpened(issueNum, issueURL)
	case "closed":
		ex = story.OnReviewRequestClosed(issueNum, issueURL)
	case "reopened":
		ex = story.OnReviewRequestReopened(issueNum, issueURL)
	default:
		panic("unreachable code reached")
	}
	if ex != nil {
		httputils.Error(rw, r, err)
		return
	}

	if *event.Action == "closed" {
		if err := story.MarkAsReviewed(); err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}

	httpStatus(rw, http.StatusAccepted)
}
