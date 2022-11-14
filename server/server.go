// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package server contains functionality for running a local server to help
// guide the user through a SAML login.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Start an HTTP server listening which guides the user through a SAML login.
// Returns a function that must be invoked by the caller to wait for the SAML
// response and the server to shutdown.
func Start(ctx context.Context, listen, url string) (string, func() (string, error)) { //nolint:funlen
	// Channels for asynchronously communicating the SAML response string or
	// any errors that are encountered.
	responseChan := make(chan string)
	errChan := make(chan error)

	// sendError is a helper function for sending errors to the error channel
	// in a non-blocking fashion.
	sendError := func(err error) {
		if err != nil {
			go func() {
				errChan <- err
			}()
		}
	}

	ctx, cancel := context.WithCancel(ctx)

	mux := http.NewServeMux()

	// Serve frontend web app content.
	mux.Handle("/", static)

	// The /login route redirects the user to the IdP initiated login url.
	mux.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
	})

	// The /callback route receives the POST which contains the SAML response.
	mux.HandleFunc("/callback", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		if err := request.ParseForm(); err != nil {
			http.Error(writer, "bad request", http.StatusBadRequest)
			sendError(errors.New("failed to parse post form"))

			return
		}

		go func() {
			// Write the SAML response string to our channel.
			responseChan <- request.FormValue("SAMLResponse")
		}()

		http.Redirect(writer, request, "/", http.StatusFound)
	})

	// The /callback route is called by the user to terminate this server.
	mux.HandleFunc("/shutdown", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		// Cancel the underlying context.
		cancel()
	})

	// Wait for the underlying context to expire of be cancelled and forward
	// any errors.
	go func() {
		<-ctx.Done()
		sendError(ctx.Err())
	}()

	// Create a custom http.Server so that it can be manually shutdown.
	server := &http.Server{ //nolint:exhaustruct
		Addr:              listen,
		Handler:           mux,
		ReadHeaderTimeout: time.Minute,
	}

	// Start the http server and forward any errors.
	go func() {
		sendError(server.ListenAndServe())
	}()

	// Return a function that can be used by the caller to wait for either the
	// SAML response of an error.
	return fmt.Sprintf("http://%s/login", listen), func() (string, error) {
		defer server.Shutdown(ctx) //nolint:errcheck

		var samlResponse string
		// Wait for the SAML response.
		go func() {
			samlResponse = <-responseChan
		}()

		// Wait for an error.
		err := <-errChan
		switch err {
		case context.Canceled:
			// The context being cancelled is an explicit action and not
			// indicative of a runtime error. Check to see if a SAML response
			// was already received.
			if samlResponse != "" {
				return samlResponse, nil
			}

			return "", errors.New("saml response was not received")
		case context.DeadlineExceeded:
			// The user too long to login to the SAML IdP.
			return "", errors.New("timeout exceeded")
		default:
			return "", err
		}
	}
}
