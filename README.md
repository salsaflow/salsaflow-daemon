# salsaflow-daemon

A simple daemon that handles asynchronous events for SalsaFlow.

The daemon is designed to run on Heroku, but it can be deployed anywhere.

## Endpoints

Each endpoint represents certain piece of functionality that you might want to
enable for your SalsaFlow-enabled project.

### `/issuetracker/pivotaltracker/events`

Server-side counterpart for the Pivotal Tracker issue tracker module.

#### Setup

Required environment variables:

* `PIVOTALTRACKER_TOKEN`
* `PIVOTALTRACKER_REVIEWED_LABEL`
* `PIVOTALTRACKER_TESTING_PASSED_LABEL`
* `PIVOTALTRACKER_TESTING_FAILED_LABEL`
* `PIVOTALTRACKER_IMPLEMENTED_LABEL`

### `/codereview/github/events`

Server-side counterpart for the GitHub code review module.

#### Setup

Required environment variables:

* `GITHUB_TOKEN`

Optional environment variables:

* `GITHUB_SECRET`

In case you are using Pivotal Tracker as the issue tracker:

* `PIVOTALTRACKER_TOKEN`

In case you are using JIRA as the issue tracker:

* `JIRA_BASE_URL`
* `JIRA_OAUTH_ACCESS_TOKEN`
* `JIRA_OAUTH_CONSUMER_KEY`
* `JIRA_OAUTH_PRIVATE_KEY`
