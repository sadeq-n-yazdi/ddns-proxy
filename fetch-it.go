package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
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
	Password    string `json:"password,omitempty,default:'<PASSWORD>'"`
	UserID      string `json:"id,omitempty,default:'demo'"`
	Host        string `json:"host,omitempty,default:'example.com'"`
	DDUser      string `json:"dd-user,omitempty,default:'demo'"`
	DDPass      string `json:"dd-pass,omitempty,default:''"`
	UrlPattern  string `json:"url,omitempty,default:''"`
	ForceUpdate bool   `json:"force-update,omitempty,default:false"`
}

const (
	defaultURLPattern = "https://{dduser}:{ddpass}@domains.google.com/nic/update?hostname={ddhost}&myip={ddip}"
)

var copyrightParameters = map[string]interface{}{
	"year":               time.Now().Format("2006"),
	"author":             "Sadeq N. Yazdi",
	"name":               "Sadeq N. Yazdi",
	"email":              "codes@sadeq.uk",
	"url":                "https://github.com/sadeq-n-yazd/go-ddns",
	"license":            "GPLv3",
	"applicationExeName": "DDNS-Proxy",
}

// Define a map to store user credentials
var validCredentials = make(map[string]UserInfo)

func main() {
	printCopyright(false)

	cfg = getConfig("./")
	if cfg.Debug {
		getLogger().SetLevel(logrus.DebugLevel)
	}
	for i := 1; i < len(os.Args); i++ {
		if strings.TrimSpace(strings.ToLower(os.Args[i])) == "-cc" {
			printCopyright(true)
			return
		}
	}

	if err := setupCredentialsFromFile(); err != nil {
		getLogger().WithError(err).Fatal("Failed to setup credentials from file")
	}

	http.HandleFunc("/", fetchItHandlerFunc)
	http.HandleFunc("/about", AboutHandlerFunc)

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
			} else if field == "force-update" {
				forceUpdate, _ := strconv.ParseBool(val.String())
				creds.ForceUpdate = forceUpdate
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

	theUrl := ""
	getLogger().Debug("requestedIpAddress:", requestedIpAddress)
	pattern := defaultURLPattern
	if checkValidAPICredentials(creds) && requestedIpAddress != "" {
		if creds.UrlPattern != "" {
			pattern = creds.UrlPattern
		}

		theUrl = Interpolate(pattern, map[string]interface{}{
			"ddhost": url.QueryEscape(creds.Host),
			"dduser": url.QueryEscape(creds.DDUser),
			"ddpass": url.QueryEscape(creds.DDPass),
			"ddip":   url.QueryEscape(requestedIpAddress),
		})

		if !creds.ForceUpdate {
			ips, err := net.LookupIP(creds.Host)
			if err == nil && len(ips) > 0 {
				if ips[0].String() == requestedIpAddress {
					getLogger().Infof("IP %s already is set for %s", requestedIpAddress, creds.Host)
					w.Header().Set("Content-Type", "text/plain")
					_, _ = w.Write([]byte("nochn : IP already is set "))
					_, _ = w.Write([]byte(requestedIpAddress))
					return
				}
			}
		}

		getLogger().Debugf("%v", []string{"URL", theUrl})

		if cfg.Debug {
			time.Sleep(10 * time.Second)
		}
		// Create an HTTP client with a timeout of 5 seconds
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		paramsMap, err := GetParamsAsMap(r)
		theUrl = Interpolate(theUrl, *paramsMap)
		// Send an HTTP GET request using the client
		getLogger().Debugf("Compiled URL using parameters: %v", []string{"URL", theUrl})
		resp, err := client.Get(theUrl)
		if err != nil {
			getLogger().Warn("Error making GET request:", err)
			http.Error(w, "Error making GET request", http.StatusBadRequest)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			getLogger().Warn("Error reading response body", err)
			http.Error(w, "Error reading response body", http.StatusBadGateway)
			return
		}

		getLogger().Debug("Body: ", string(body))
		getLogger().Debug("Header: ", resp.Header)
		// Check if the response body contains "ok"

		if resp.StatusCode == http.StatusOK {
			getLogger().Infof("Respons: %d\n%s", resp.StatusCode, body)
			_, _ = fmt.Fprintf(w, "OK %d\n%s", resp.StatusCode, body)
		} else {
			getLogger().Warnf("Respons: %d\n%s", resp.StatusCode, body)
			_, _ = fmt.Fprintf(w, "Fail: %d\n%s", resp.StatusCode, body)
		}

	} else {
		getLogger().Warn("Credentials are not valid, ", creds)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
}

//go:embed copyright-banner.txt
var embeddedCopyrightBanner string

//go:embed copyrigth-full.txt
var embeddedCopyright string

//go:embed licence.md
var embeddedLicence string

func printCopyright(full bool) {
	var c string
	appName, _ := os.Executable()
	appName = filepath.Base(appName)
	copyrightParameters["applicationExeName"] = appName

	if full {
		c = Interpolate(embeddedCopyright, copyrightParameters)
	} else {
		c = Interpolate(embeddedCopyrightBanner, copyrightParameters)
	}

	c = Interpolate(c, copyrightParameters)
	getLogger().Info(c)
}

func AboutHandlerFunc(w http.ResponseWriter, _ *http.Request) {
	appName, _ := os.Executable()
	appName = filepath.Base(appName)
	copyrightParameters["applicationExeName"] = appName

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	cc := Interpolate(embeddedCopyright, copyrightParameters)
	_, _ = w.Write([]byte(cc))
	_, _ = w.Write([]byte("\n\n" + strings.Repeat("-", 79) + "\n\n"))
	_, _ = w.Write([]byte(embeddedLicence))
}
