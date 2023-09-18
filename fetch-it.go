package main

import (
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// UserInfo Define a custom struct type with JSON tags and default values
type UserInfo struct {
	Password string `json:"password,omitempty,default:'<PASSWORD>'"`
	UserID   string `json:"id,omitempty,default:'demo'"`
	Host     string `json:"host,omitempty,default:'example.com'"`
	DDUser   string `json:"dd-user,omitempty,default:'demo'"`
	DDPass   string `json:"dd-pass,omitempty,default:''"`
}

// Define a map to store user credentials
var validCredentials = make(map[string]UserInfo)

func main() {
	cfg = getConfig("./")
	if cfg.Debug {
		getLogger().SetLevel(logrus.DebugLevel)
	}

	if err := setupCredentialsFromFile(); err != nil {
		getLogger().WithError(err).Fatal("Failed to setup credentials from file")
	}

	http.HandleFunc("/", fetchItHandlerFunc)

	// Start the HTTP server
	port := cfg.HostName + ":" + strconv.Itoa(cfg.Port)
	getLogger().Infof("Starting server on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		getLogger().Error("Error starting server:", err)
	}
}

func setupCredentialsFromFile() error {
	// Try to read credentials from different locations
	wd, _ := os.Getwd()
	wd, _ = filepath.Abs(wd)

	locations := []string{
		"/etc/websites/fetchit.sadeq.uk/cred.jsonc",   // First, check in /etc/websites/fetch-it.sadeq.uk/
		"/etc/fetchit/cred.jsonc",                     // First, check in /etc/fetch-it/
		filepath.Join(wd, ".cred.jsonc"),              // Then, check in the current working directory
		filepath.Join(executableDir(), ".cred.jsonc"), // Then, check beside the executable
	}

	var credentialFile string

	for _, location := range locations {
		getLogger().Debug("checking credential file at: ", location)
		if _, err := os.Stat(location); err == nil {
			credentialFile = location
			break
		}
	}

	// Check if a credential file was found
	if credentialFile == "" {
		getLogger().Error("credential file not found in expected locations.")
		return fmt.Errorf("credential file not found in expected locations")
	}

	// Read credentials from the JSONC file
	if err := readCredentialsFromFile(credentialFile); err != nil {
		getLogger().Error("error reading credentials:", err)
		return fmt.Errorf("error reading credentials: %s", err)
	}
	return nil
}

// Get the directory of the currently executing executable
func executableDir() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	getLogger().Debug("Executable file:", executable)
	getLogger().Debug("Executable path:", filepath.Dir(executable))
	return filepath.Dir(executable)
}

// Read credentials from a JSONC file and populate the validCredentials map
func readCredentialsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		getLogger().Error("Can not read credential file:", err)
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
		creds := UserInfo{}
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

func checkValidAPICredentials(user *UserInfo) bool {
	userExists := strings.TrimSpace(user.DDUser) != ""
	passExists := strings.TrimSpace(user.DDPass) != ""
	getLogger().Debug("User exists:", userExists, []string{user.DDUser})
	getLogger().Debug("Pass exists:", passExists, []string{user.DDPass})

	return userExists && passExists
}

func authorize(w http.ResponseWriter, r *http.Request) (creds *UserInfo, ok bool) {
	creds = &UserInfo{}
	//Get the Authorization header from the request
	authHeader := r.Header.Get("Authorization")

	// Check if the Authorization header is not empty and starts with "Basic "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
		// Authorization header is missing or not in the correct format
		getLogger().Info("Unknown authorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	// Extract the base64-encoded credentials (after "Basic ")
	authValue := authHeader[len("Basic "):]

	// Decode the base64-encoded credentials
	credentials, err := base64.StdEncoding.DecodeString(authValue)
	if err != nil {
		getLogger().Warn("Bad encoding for credentials")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	// Split the decoded credentials into a username and password
	parts := strings.SplitN(string(credentials), ":", 2)
	if len(parts) != 2 {
		getLogger().Warn("Empty credentials")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	// Extract the username and password
	username := parts[0]
	password := parts[1]

	getLogger().Debug("Header: username:", username, " password:", password)

	// Check if the provided credentials are valid
	var exists bool
	*creds, exists = validCredentials[username]
	if !exists || password != creds.Password {
		if !exists {
			getLogger().Warn("Creds: username:", username, "not found!")
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}
	return creds, true
}

func getRealIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	ip = strings.SplitN(ip, ":", 2)[0]
	return ip
}
func fetchItHandlerFunc(w http.ResponseWriter, r *http.Request) {
	creds, valid := authorize(w, r)
	if !valid {
		// if auth fail not handled in authorize func
		getLogger().Debug("auth failed")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Authorization failed"))
		return
	}

	requestedIpAddress := r.URL.Query().Get("ip")
	if requestedIpAddress == "" {
		requestedIpAddress = getRealIP(r)
	}

	url := ""
	getLogger().Debug("requestedIpAddress:", requestedIpAddress)
	if checkValidAPICredentials(creds) && requestedIpAddress != "" {
		url = fmt.Sprintf(
			"https://%s:%s@domains.google.com/nic/update?hostname=%s&myip=%s",
			creds.DDUser,
			creds.DDPass,
			creds.Host,
			requestedIpAddress,
		)
		getLogger().Debugf("%v", []string{"URL", url})

		if cfg.Debug {
			time.Sleep(10 * time.Second)
		}
		// Create an HTTP client with a timeout of 5 seconds
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		// Send an HTTP GET request using the client
		resp, err := client.Get(url)
		if err != nil {
			getLogger().Warn("Error making GET request:", err)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			getLogger().Warn("Error reading response body", err)
			return
		}

		getLogger().Debug("Body: ", body)
		getLogger().Debug("Header: ", resp.Header)
		// Check if the response body contains "ok"
		if strings.Contains(string(body), "ok") {
			getLogger().Info("Response contains 'ok'")
		} else {
			getLogger().Info("Response does not contain 'ok'")
		}

	} else {
		getLogger().Warn("Credentials are not valid, ", creds)
	}
	response := url
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, response)

}
