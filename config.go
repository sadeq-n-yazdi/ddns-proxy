package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"os"
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
	sectionName     = "server"
)

func getConfig(path string) *ServerConfig {
	var configFileName string
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
	println("config file: ", configFileName)
	settings, err := ini.Load(configFileName)
	if err != nil {
		return &defaultConfig
	}

	if debug, err := settings.Section(sectionName).Key("debug").Bool(); err == nil {
		defaultConfig.Debug = debug
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
	var fName *string
	// Get the path of the executable file
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeName := *stripExtension(filepath.Base(exePath))
	if exeName == "" {
		println("Exe path is", exePath, "and exeName is", exeName, "!")
	}
	// Get the directory of the executable file
	exeDir := filepath.Dir(exePath)

	// println(exePath, exeDir, exeName)
	fName = concatFileNames("/etc/websites/"+exeName, defaultFileName)
	println("hopeful to find config file in:", *fName)
	if fileIsReadable(fName) {
		println(*fName)
		return *fName, nil
	}

	fName = concatFileNames(exeDir, defaultFileName)
	if fileIsReadable(fName) {
		println(*fName)
		return *fName, nil
	}

	fName = concatFileNames(exeDir, exeName+".ini")
	if fileIsReadable(fName) {
		println(*fName)
		return *fName, nil
	}
	println("config file not found")
	return "", fmt.Errorf("could not find config file")
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

func concatFileNames(path string, filename string) *string {
	fullFileName := path + "/" + filename
	return &fullFileName
}
