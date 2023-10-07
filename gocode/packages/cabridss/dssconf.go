package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"os"
	"time"
)

type DssBaseConfig struct {
	ConfigDir         string                                                      `json:"-"`          // if not "" path to the user's configuration directory
	ConfigPassword    string                                                      `json:"-"`          // if master password is used to encrypt client configuration
	LocalPath         string                                                      `json:"-"`          // local path for configuration and index, or "" if unused (index will be memory based)
	RepoId            string                                                      `json:"repoId"`     // uuid of the repository
	Unlock            bool                                                        `json:"-"`          // unlocks index concurrent updates lock
	AutoRepair        bool                                                        `json:"autoRepair"` // if unlock required, automatically repairs the index
	ReIndex           bool                                                        `json:"-"`          // forces full content reindexation
	GetIndex          func(config DssBaseConfig, localPath string) (Index, error) `json:"-"`          // non-default function to instantiate an index
	XImpl             string                                                      `json:"xImpl"`      // index implementation code: bdb, memory, no
	LibApi            bool                                                        `json:"-"`          // prevents using a web API server for local DSS access
	WebProtocol       string                                                      `json:"-"`          // web API server protocol
	WebHost           string                                                      `json:"-"`          // web API server host
	WebPort           string                                                      `json:"-"`          // web API server port
	WebClientTimeout  time.Duration                                               `json:"-"`          // client timeout in seconds, a Timeout of zero means no timeout
	TlsCert           string                                                      `json:"-"`          // certificate file on https server or untrusted CA on https client
	TlsKey            string                                                      `json:"-"`          // certificate key file on https server
	TlsNoCheck        bool                                                        `json:"-"`          // no check of certificate by https client
	BasicAuthUser     string                                                      `json:"-"`          // adds basic authentication
	BasicAuthPassword string                                                      `json:"-"`          // basic authentication password
	WebRoot           string                                                      `json:"-"`          // web API server root
	Encrypted         bool                                                        `json:"encrypted"`  // repository is encrypted
	ReducerLimit      int                                                         `json:"-"`          // if not 0 max number of parallel I/O
}

func writeDssConfig(bc DssBaseConfig, dssConfig interface{}) error {
	if err := checkDir(bc.LocalPath); err != nil {
		return fmt.Errorf("in SaveDssConfig: %w", err)
	}
	configFile := ufpath.Join(bc.LocalPath, "config")
	bs, err := json.Marshal(dssConfig)
	if err != nil {
		return fmt.Errorf("in SaveDssConfig: %w", err)
	}
	uc, err := CurrentUserConfig(bc)
	if err != nil {
		return fmt.Errorf("in SaveDssConfig: %w", err)
	}
	ebs, err := EncryptMsg(string(bs), uc.Internal.PKey)
	if err != nil {
		return fmt.Errorf("in SaveDssConfig: %w", err)
	}
	if err := os.WriteFile(configFile, ebs, 0o666); err != nil {
		return fmt.Errorf("in SaveDssConfig: %w", err)
	}
	return nil
}

func SaveDssConfig(bc DssBaseConfig, dssConfig interface{}) error {
	configFile := ufpath.Join(bc.LocalPath, "config")
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("in SaveDssConfig: cannot create file %s", configFile)
	}
	return writeDssConfig(bc, dssConfig)
}

func OverwriteDssConfig(bc DssBaseConfig, dssConfig interface{}) error {
	configFile := ufpath.Join(bc.LocalPath, "config")
	if _, err := os.Stat(configFile); err != nil {
		return fmt.Errorf("in SaveDssConfig: cannot load file %s", configFile)
	}
	return writeDssConfig(bc, dssConfig)
}

func LoadDssConfig(bc DssBaseConfig, persistentConfig interface{}) error {
	ebs, err := os.ReadFile(ufpath.Join(bc.LocalPath, "config"))
	if err != nil {
		return fmt.Errorf("in LoadDssConfig: %w", err)
	}
	uc, err := CurrentUserConfig(bc)
	if err != nil {
		return fmt.Errorf("in LoadDssConfig: %w", err)
	}
	bs, err := DecryptMsg(ebs, uc.Internal.Secret)
	if err != nil {
		return fmt.Errorf("in LoadDssConfig: %w", err)
	}
	if err = json.Unmarshal([]byte(bs), persistentConfig); err != nil {
		return fmt.Errorf("in LoadDssConfig: %w", err)
	}
	return nil
}
