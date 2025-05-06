package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	Penalty     int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

func GetConfig() (*Config, error) {
	var cfg Config
	file, err := os.Open("config/config.json")
	if err != nil {
		return nil, fmt.Errorf("error opening file %v", err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file %v", err)
	}
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %v", err)
	}
	return &cfg, nil
}
