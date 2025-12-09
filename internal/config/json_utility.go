package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func Read() (Config, error) {
	var currentConfig Config
	userhome, osError := os.UserHomeDir()
	if osError != nil {
		return Config{}, fmt.Errorf("Error occurred loading the HOME directory", osError)
	}
	hostjson := userhome + "/.gatorconfig.json"
	fmt.Printf("Current directory and file: %s\n", hostjson)
	jsonFile, err := os.ReadFile(hostjson)
	if err != nil {
		return Config{}, fmt.Errorf("Error occurred loading file or file not found: ", err)
	}

	if err2 := json.Unmarshal(jsonFile, &currentConfig); err2 != nil {
		return Config{}, fmt.Errorf("Unable to read", err2)
	}

	return currentConfig, nil
}

func SetUser(currentConfig Config, user string) error {
	currentConfig.CurrentUserNname = user
	write(currentConfig)
	return nil
}

func write(cfg Config) error {
	userhome, osError := os.UserHomeDir()
	if osError != nil {
		return fmt.Errorf("Error occurred loading the HOME directory", osError)
	}
	hostjson := userhome + "/.gatorconfig.json"

	preppedJson, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Error in building json: ", err)
	}

	writeError := os.WriteFile(hostjson, preppedJson, 0o664)
	if writeError != nil {
		return fmt.Errorf("Error writing: ", writeError)
	}
	return nil
}
