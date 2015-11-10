.PHONY: install test godep_test

TEST=go test -v
GODEP_TEST= godep go test -v

install:
	go install github.com/salsaflow/salsaflow-daemon

test: CMD=go test -v
test: internal.test

godep-test: CMD=godep go test -v
godep-test: internal.test

internal.test:
	${CMD} \
		github.com/salsaflow/salsaflow-daemon/internal/modules/codereview/github/endpoint \
		github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/endpoint \
		github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/tracker
