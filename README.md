# salsaflow-daemon

A simple daemon that handles asynchronous events for SalsaFlow.

The daemon is designed to run on Heroku, but it can be deployed anywhere.

## Significant Environment Variables

The following environment variables can be used to configure `salsaflow-daemon`.

### Pivotal Tracker-related Logic

* `PIVOTALTRACKER_ACCESS_TOKEN` - Pivotal Tracker access token
* `PIVOTALTRACKER_SECRET` - secret used to authenticate incoming webhooks
* `PIVOTALTRACKER_REVIEWED_LABEL` - the label marking PT stories as reviewed
* `PIVOTALTRACKER_SKIP_REVIEW_LABEL` - the label saying that PT story doesn't need review
* `PIVOTALTRACKER_PASSED_TESTING_LABEL` - the label marking PT story as passing QA
* `PIVOTALTRACKER_FAILED_TESTING_LABEL` - the label marking PT story as failing QA
* `PIVOTALTRACKER_SKIP_TESTING_LABEL` - the label saying that PT story doesn't need QA

### JIRA-related Logic

* `JIRA_API_BASE_URL` - JIRA API address, e.g. `https://jira.example.com/rest/api/2/`
* `JIRA_OAUTH_ACCESS_TOKEN` - JIRA OAuth access token
* `JIRA_OAUTH_CONSUMER_KEY` - JIRA OAuth consumer key
* `JIRA_OAUTH_PRIVATE_KEY` - JIRA OAuth RSA private key

### GitHub-related Logic

* `GITHUB_ACCESS_TOKEN` - token to be used when calling GitHub API
* `GITHUB_WEBHOOK_SECRET` - secret used to authenticate incoming webhooks

## Endpoints

Each endpoint represents certain piece of functionality that you might want to
enable for your SalsaFlow-enabled project. To enable the functionality you simply
need to point relevant webhook to the right endpoint.

### `/issuetracker/pivotaltracker/events`

Server-side counterpart of Salsita's Pivotal Tracker issue tracker module.

This endpoint performs the following actions:

* When a rejected story is detected, it remotes all review and QA labels.

This endpoint expects Pivotal Tracker `v5` activity webhooks.

#### Setup

Required environment variables:

* `PIVOTALTRACKER_ACCESS_TOKEN`

Optional environment variables:

* `PIVOTALTRACKER_SECRET` (highly recommended to set this variable)
* `PIVOTALTRACKER_REVIEWED_LABEL` (default `reviewed`)
* `PIVOTALTRACKER_SKIP_REVIEW_LABEL` (default `no review`)
* `PIVOTALTRACKER_PASSED_TESTING_LABEL` (default `qa+`)
* `PIVOTALTRACKER_FAILED_TESTING_LABEL` (default `qa-`)
* `PIVOTALTRACKER_SKIP_TESTING_LABEL` (default `no qa`)

In case you decide to set `PIVOTALTRACKER_SECRET`, append `?secret=<secret-value>`
to the webhook destination URL. Incoming webhooks missing this parameter
will be automatically rejected.

### `/codereview/github/events`

Server-side counterpart of Salsita's GitHub code review module.

This endpoint performs the following actions:

* When a GitHub review issue is closed, it marks relevant issue as reviewed.
  What that means depends on the associated issue tracker.
* When the review issue is re-opened again, it resets the associated story
  back to the state representing the fact that the story is being reviewed.
* It processes commands in GitHub commit comments, particularly `!blocker`.

#### Setup

Required environment variables:

* `GITHUB_ACCESS_TOKEN`

Optional environment variables:

* `GITHUB_WEBHOOK_SECRET` (highly recommended to set this variable)

In case you are using Pivotal Tracker as the issue tracker,
the following variables are required:

* `PIVOTALTRACKER_TOKEN`

In case you are using JIRA as the issue tracker,
the following variables are required:

* `JIRA_API_BASE_URL`
* `JIRA_OAUTH_ACCESS_TOKEN`
* `JIRA_OAUTH_CONSUMER_KEY`
* `JIRA_OAUTH_PRIVATE_KEY`
