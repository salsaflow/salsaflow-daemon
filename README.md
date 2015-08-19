# salsaflow-daemon

A simple daemon that handles asynchronous events for SalsaFlow.

The daemon is designed to run on Heroku, but it can be deployed anywhere.

## Significant Environment Variables

The following environment variables can be used to configure `salsaflow-daemon`.

### Pivotal Tracker-related Logic

* `PIVOTALTRACKER_ACCESS_TOKEN` - Pivotal Tracker access token
* `PIVOTALTRACKER_REVIEWED_LABEL` - the label used to mark PT stories as reviewed
* `PIVOTALTRACKER_SKIP_REVIEW_LABEL` - the label used to say that the PT story doesn't need review
* `PIVOTALTRACKER_PASSED_TESTING_LABEL` - the label used to mark PT story as passing QA
* `PIVOTALTRACKER_FAILED_TESTING_LABEL` - the label used to mark PT story as failing QA
* `PIVOTALTRACKER_SKIP_TESTING_LABEL` - the label used to say that the PT story doesn't need QA

### JIRA-related Logic

* `JIRA_BASE_URL` - JIRA API address, e.g. `https://jira.example.com/rest/api/2/`
* `JIRA_OAUTH_ACCESS_TOKEN` - JIRA OAuth access token
* `JIRA_OAUTH_CONSUMER_KEY` - JIRA OAuth consumer key
* `JIRA_OAUTH_PRIVATE_KEY` - JIRA OAuth RSA private key

### GitHub-related Logic

* `GITHUB_ACCESS_TOKEN` - token to be used when calling GitHub API
* `GITHUB_WEBHOOK_SECRET` - secret used to authenticate incoming webhooks

## Endpoints

Each endpoint represents certain piece of functionality that you might want to
enable for your SalsaFlow-enabled project.

### `/issuetracker/pivotaltracker/events`

Server-side counterpart of Salsita's Pivotal Tracker issue tracker module.

#### Setup

Required environment variables:

* `PIVOTALTRACKER_ACCESS_TOKEN`

Optional environment variables:

* `PIVOTALTRACKER_REVIEWED_LABEL` (default `reviewed`)
* `PIVOTALTRACKER_SKIP_REVIEW_LABEL` (default `no review`)
* `PIVOTALTRACKER_PASSED_TESTING_LABEL` (default `qa+`)
* `PIVOTALTRACKER_FAILED_TESTING_LABEL` (default `qa-`)
* `PIVOTALTRACKER_SKIP_TESTING_LABEL` (default `no qa`)

### `/codereview/github/events`

Server-side counterpart of Salsita's GitHub code review module.

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

* `JIRA_BASE_URL`
* `JIRA_OAUTH_ACCESS_TOKEN`
* `JIRA_OAUTH_CONSUMER_KEY`
* `JIRA_OAUTH_PRIVATE_KEY`
