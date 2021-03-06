package github

import (
	// Stdlib
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	// Internal
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"

	// Vendor
	"github.com/codegangsta/negroni"
	"github.com/google/go-github/github"
)

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Read the request body into a buffer.
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				httputil.Error(rw, r, err)
				return
			}

			// Fill the request body again so that it is available in the next handler.
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			// Compute the hash.
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := io.Copy(mac, bytes.NewReader(bodyBytes)); err != nil {
				httputil.Error(rw, r, err)
				return
			}

			// Compare with the header provided in the request.
			secretHeader := r.Header.Get("X-Hub-Signature")
			expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
			if secretHeader != expected {
				repo := getRepoFullName(bodyBytes)
				log.Warn(r, "HMAC mismatch detected (repo='%v'): expected='%v', got='%v'\n",
					repo, expected, secretHeader)
				httputil.Status(rw, http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}

func getRepoFullName(body []byte) string {
	var payload github.WebHookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return *payload.Repo.FullName
}
