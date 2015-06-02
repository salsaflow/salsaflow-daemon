/*
   Copyright (C) 2014  Salsita s.r.o.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program. If not, see {http://www.gnu.org/licenses/}.
*/

package jira

import (
	// Stdlib
	"net/http"

	// Vendor
	"github.com/tchap/oauth"
)

func NewOAuthClient(clientKey, clientSecret, accessToken string) *http.Client {
	return &http.Client{
		Transport: newOAuthRoundTripper(clientKey, clientSecret, accessToken),
	}
}

type oauthRoundTripper struct {
	consumer *oauth.Consumer
	token    *oauth.AccessToken
}

func newOAuthRoundTripper(clientKey, clientSecret, accessToken string) *oauthRoundTripper {
	return &oauthRoundTripper{
		consumer: oauth.NewConsumer(clientKey, clientSecret, oauth.ServiceProvider{}),
		token:    &oauth.AccessToken{Token: accessToken},
	}
}

func (rt *oauthRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.consumer.MakeRequest(r, rt.token)
}
