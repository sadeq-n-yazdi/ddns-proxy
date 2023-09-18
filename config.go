package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
)

type ServerConfig struct {
	Port     int
	HostName string
	CertFile string
	KeyFile  string
	CAFile   string
	CAPath   string
	SSL      bool
	Debug    bool
}

var cfg *ServerConfig

const (
	defaultFileName = "config.ini"
)

func getConfig(path string) *ServerConfig {
	var configFileName string
	sectionName := ini.DEFAULT_SECTION

	// Define default config values
	defaultConfig :=
		ServerConfig{
			Port:     443,
			HostName: "localhost",
			CertFile: "server.crt",
			KeyFile:  "server.key",
			SSL:      false,
			Debug:    true,
		}
	if fileIsReadable(&path) {
		configFileName = path
	} else {
		cfgFileName, err := getConfigFilePath()
		if err != nil {
			return &defaultConfig
		}
		configFileName = cfgFileName
	}
	getLogger().Info("config file: ", configFileName)
	settings, err := ini.Load(configFileName)
	if err != nil {
		return &defaultConfig
	}

	if debug, err := settings.Section(sectionName).Key("debug").Bool(); err == nil {
		defaultConfig.Debug = debug
		if debug {
			getLogger().SetLevel(logrus.DebugLevel)
			getLogger().Debug("Debug mode enabled")
		}
	}
	defaultConfig.HostName = settings.Section(sectionName).Key("host").String()

	if port, err := settings.Section(sectionName).Key("port").Int(); (err == nil) && (port > 0) {
		defaultConfig.Port = port
	}
	if certFile := settings.Section(sectionName).Key("cert").String(); certFile != "" {
		defaultConfig.CertFile = certFile
	}
	if keyFile := settings.Section(sectionName).Key("key").String(); keyFile != "" {
		defaultConfig.KeyFile = keyFile
	}
	if ssl, err := settings.Section(sectionName).Key("secure").Bool(); err == nil {
		defaultConfig.SSL = ssl
	}

	return &defaultConfig
}

func getConfigFilePath() (string, error) {
	var (
		exeName         string
		exeDir          string
		configFilePaths []string
	)

	// Get the path of the executable file
	exePath, err := os.Executable()
	if err == nil {
		exeName = *stripExtension(filepath.Base(exePath))
		if exeName == "" {
			getLogger().Debug("Exe path is", exePath, "and exeName is", exeName, "!")
		}
		// Get the directory of the executable file
		exeDir = filepath.Dir(exePath)
	}

	configFilePaths = append(configFilePaths, path.Join("/etc", "websites", exeName, defaultFileName))
	configFilePaths = append(configFilePaths, path.Join(exeDir, defaultFileName))
	wd, err := os.Getwd()
	if err == nil && wd != exeDir {
		configFilePaths = append(configFilePaths, path.Join(wd, defaultFileName))
		configFilePaths = append(configFilePaths, path.Join(wd, exeName+".ini"))
	}
	configFilePaths = append(configFilePaths, path.Join(exeDir, exeName+".ini"))
	configFilePaths = append(configFilePaths, path.Join("/etc", "websites", "fetchit.sadeq.uk", defaultFileName))

	for _, fn := range configFilePaths {
		if fileIsReadable(&fn) {
			getLogger().Debug("Config file found at", fn)
			return fn, nil
		}
	}
	getLogger().Error("config file not found!")
	return "", fmt.Errorf("no config file found")
}

func fileIsReadable(filename *string) bool {

	info, err := os.Stat(*filename)
	if err != nil {
		print(err)
		return false
	}

	if info.IsDir() {
		return false
	}

	f, err := os.OpenFile(*filename, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

func stripExtension(filename string) *string {
	ext := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(ext)]
	return &filename
}
