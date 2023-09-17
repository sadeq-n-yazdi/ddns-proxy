package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Define a map to store user credentials
var validCredentials = make(map[string]struct {
	Password string `json:"password,omitempty,default:'<PASSWORD>'"`
	UserID   string `json:"user,omitempty,default:'demo'"`
	Host     string `json:"host,omitempty,default:'example.com'"`
	DDUser   string `json:"dd-user,omitempty,default:'demo'"`
	DDPass   string `json:"dd-pass,omitempty,default:''"`
})

func main() {
	// Try to read credentials from different locations
	wd, _ := os.Getwd()
	locations := []string{
		"/etc/fetch-it/cred.jsonc",                    // First, check in /etc/fetch-it/
		filepath.Join(wd, ".cred.jsonc"),              // Then, check in the current working directory
		filepath.Join(executableDir(), ".cred.jsonc"), // Then, check beside the executable
	}

	var credentialFile string

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			credentialFile = location
			break
		}
	}

	// Check if a credential file was found
	if credentialFile == "" {
		fmt.Println("Credential file not found in expected locations.")
		return
	}

	// Read credentials from the JSONC file
	if err := readCredentialsFromFile(credentialFile); err != nil {
		fmt.Println("Error reading credentials:", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // Get the Authorization header from the request
		authHeader := r.Header.Get("Authorization")

		// Check if the Authorization header is not empty and starts with "Basic "
		if authHeader != "" && strings.HasPrefix(authHeader, "Basic ") {
			// Extract the base64-encoded credentials (after "Basic ")
			authValue := authHeader[len("Basic "):]

			// Decode the base64-encoded credentials
			credentials, err := base64.StdEncoding.DecodeString(authValue)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Split the decoded credentials into username and password
			parts := strings.SplitN(string(credentials), ":", 2)
			if len(parts) != 2 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract the username and password
			username := parts[0]
			password := parts[1]

			// Check if the provided credentials are valid
			creds, exists := validCredentials[username]
			if !exists || password != creds.Password {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			requestedIpAddress := r.URL.Query().Get("ip")
			if requestedIpAddress == "" {
				requestedIpAddress = r.RemoteAddr
			}
			if r.Header.Get("X-Forwarded-For") != "" {
				requestedIpAddress = r.Header.Get("X-Forwarded-For")
			}
			if r.Header.Get("X-Real-Ip") != "" {
				requestedIpAddress = r.Header.Get("X-Real-Ip")
			}
			// Authentication passed, respond with "OK" and include additional fields
			//response := fmt.Sprintf("OK, User ID: %s, Host: %s, DDUser: %s, DDPass: %s\n",
			//	creds.UserID, creds.Host, creds.DDUser, creds.DDPass)
			//fmt.Fprintln(w, response)
			response := fmt.Sprintf(
				"https://%s:%s@domains.google.com/nic/update?hostname=%s&myip=%s",
				creds.DDUser,
				creds.DDPass,
				creds.Host,
				requestedIpAddress,
			)
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			_, _ = fmt.Fprintln(w, response)
		} else {
			// Authorization header is missing or not in the correct format
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})

	// Start the HTTP server on port 9002
	// Start the HTTP server on port 9002
	port := ":9002"
	fmt.Printf("Starting server on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}

// Get the directory of the currently executing executable
func executableDir() string {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(executable)
}

// Read credentials from a JSONC file and populate the validCredentials map
func readCredentialsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Remove comments from JSONC (convert to regular JSON)
	jsonc := string(data)
	json, _ := sjson.Delete(jsonc, "//")
	json, _ = sjson.Delete(json, "#")

	// Parse JSON
	gjson.Parse(json).ForEach(func(key, value gjson.Result) bool {
		username := key.String()

		// Use default values from struct tags
		creds := struct {
			Password string `json:"password,omitempty,default:'<PASSWORD>'"`
			UserID   string `json:"user,omitempty,default:'demo'"`
			Host     string `json:"host,omitempty,default:'example.com'"`
			DDUser   string `json:"dd-user,omitempty,default:'demo'"`
			DDPass   string `json:"dd-pass,omitempty,default:''"`
		}{}

		// Unmarshal JSON data into the creds struct
		value.ForEach(func(key, val gjson.Result) bool {
			field := key.String()
			if field == "password" {
				creds.Password = val.String()
			} else if field == "user" {
				creds.UserID = val.String()
			} else if field == "host" {
				creds.Host = val.String()
			} else if field == "dd-user" {
				creds.DDUser = val.String()
			} else if field == "dd-pass" {
				creds.DDPass = val.String()
			}
			return true
		})

		validCredentials[username] = creds

		return true // Continue iterating through JSONC
	})

	return nil
}
