package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

func init() {
	currentAbPath := getCurrentAbPathByCaller()
	tomlPath, err := filepath.Abs(currentAbPath + "/config.toml")
	if err != nil {
		panic("read config.toml error: " + err.Error())
		return
	}

	_, err = UnmarshalConfig(tomlPath)
	if err != nil {
		panic("unmarshal config error: " + err.Error())
		return
	}
}

func UnmarshalConfig(configFilePath string) (*Conf, error) {
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var conf Conf
	_, err = toml.Decode(string(content), &conf)
	if err != nil {
		return nil, err
	}

	Config = &conf
	return Config, nil
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
