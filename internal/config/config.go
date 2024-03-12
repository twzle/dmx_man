package config

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type ChannelMap struct {
	SceneChannelID    uint16 `json:"scene_channel_id" yaml:"scene_channel_id"`
	UniverseChannelID uint16 `json:"universe_channel_id" yaml:"universe_channel_id"`
}

type Scene struct {
	Alias      string       `json:"scene_alias" yaml:"scene_alias"`
	ChannelMap []ChannelMap `json:"channel_map" yaml:"channel_map"`
}

type ArtNetConfig struct {
	Alias    string         `json:"alias" yaml:"alias"`
	IP       net.IP         `json:"ip" yaml:"ip"`
	Scenes   []Scene		`json:"scenes" yaml:"scenes"`
}

type DMXConfig struct {
	Alias    string         `json:"alias" yaml:"alias"`
	Path     string         `json:"path" yaml:"path"`
	Scenes   []Scene		`json:"scenes" yaml:"scenes"`
}

type UserConfig struct {
	DMXDevices    []DMXConfig    `json:"dmx_devices" yaml:"dmx_devices"`
	ArtNetDevices []ArtNetConfig `json:"artnet_devices" yaml:"artnet_devices"`
}

func (conf *UserConfig) Validate() error {
	if len(conf.DMXDevices) == 0 && len(conf.ArtNetDevices) == 0 {
		fmt.Println("DMX/ArtNet devices were not found in configuration file")
	}
	if alias, has := conf.hasDuplicateDevices("dmx"); has {
		return fmt.Errorf("found duplicate DMX device with alias {%s} in config", alias)
	}
	if alias, has := conf.hasDuplicateDevices("artnet"); has {
		return fmt.Errorf("found duplicate ArtNet device with alias {%s} in config", alias)
	}
	for idx, device := range conf.DMXDevices {
		if device.Alias == "" {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid DMX device_name must be provided in config",
				idx, device.Alias)
		}
	}
	for idx, device := range conf.ArtNetDevices {
		if device.Alias == "" {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid ArtNet device_name must be provided in config",
				idx, device.Alias)
		}
		if device.IP == nil {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid ArtNet IP address must be provided in config",
				idx, device.IP)
		}
	}
	return nil
}

func (conf *UserConfig) hasDuplicateDevices(protocol string) (string, bool) {
	x := make(map[string]struct{})

	if protocol == "dmx" {
		for _, v := range conf.DMXDevices {
			if _, has := x[v.Alias]; has {
				return v.Alias, true
			}
			x[v.Alias] = struct{}{}
		}
	} else if protocol == "artnet" {
		for _, v := range conf.ArtNetDevices {
			if _, has := x[v.Alias]; has {
				return v.Alias, true
			}
			x[v.Alias] = struct{}{}
		}
	}

	return "", false
}

func InitConfig(confPath string) (*UserConfig, error) {
	jsonFile, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	cfg, err := ParseConfigFromBytes(jsonFile)
	return cfg, err
}

func ParseConfigFromBytes(data []byte) (*UserConfig, error) {
	cfg := UserConfig{}

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
