package domoticmib

import (
	"os"

	"gopkg.in/yaml.v2"
)

type DomoticMIBAgentConfig struct {
	Device    DeviceConfig     `yaml:"device"`
	Sensors   []SensorConfig   `yaml:"sensors"`
	Actuators []ActuatorConfig `yaml:"actuators"`
}

type DomoticMIBManagerConfig struct {
	RemoteAgentsAddresses	 []string `yaml:"RemoteAgentsAddresses"`
}


func LoadMIBAgentConfig(ymlConfigPath string) (DomoticMIBAgentConfig, error) {
	bs, err := os.ReadFile(ymlConfigPath)
	if err != nil {
		return DomoticMIBAgentConfig{}, err
	}
	var config DomoticMIBAgentConfig
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		return DomoticMIBAgentConfig{}, err
	}
	return config, nil
}

func LoadMIBManagerConfig(ymlConfigPath string) (DomoticMIBManagerConfig, error) {
	bs, err := os.ReadFile(ymlConfigPath)
	if err != nil {
		return DomoticMIBManagerConfig{}, err
	}
	var config DomoticMIBManagerConfig
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		return DomoticMIBManagerConfig{}, err
	}
	return config, nil
}