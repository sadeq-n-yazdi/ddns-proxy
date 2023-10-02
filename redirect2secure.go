package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to the HTTPS URL
	http.Redirect(w, r, replacePort(r, cfg.Port), http.StatusFound)
}

func replacePort(r *http.Request, newPort int) string {
	hostParts := strings.Split(r.Host, ":")
	if len(hostParts) > 1 {
		hostParts[1] = strconv.Itoa(newPort)
	}
	newHostPort := ""
	for i, v := range hostParts {
		if i != 0 {
			newHostPort = newHostPort + ":" + v
		} else {
			newHostPort = newHostPort + v
		}
	}
	newHostPort = strings.TrimRight(newHostPort, ":")
	res := "https://" + extractUserPassword(r) + newHostPort + r.RequestURI
	return res
}

func extractUserPassword(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		// Authorization header is missing
		return ""
	}

	// Check if the header starts with “Basic “
	if !strings.HasPrefix(authHeader, "Basic ") {
		// Invalid authorization method
		return ""
	}

	// Decode the base64-encoded username:password
	encodedCreds := authHeader[6:]
	credsBytes, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		// Error decoding base64
		return ""
	}

	// Split the credentials into username and password
	creds := strings.SplitN(string(credsBytes), ":", 2)
	if len(creds) != 2 {

		return ""
	}

	username := creds[0]
	password := creds[1]

	// Now you have the username and password
	return fmt.Sprintf("%s:%s@", username, password)
}
