package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

// IdentityConfig refers to an age identity identified by an alias,
// Identities are used for encryption (PKeys of the ACL users using identities aliases)
// and for decryption (secrets of the DSS aclusers using identities aliases)
// "" is the default alias for an identity when none is provided
type IdentityConfig struct {
	Alias  string `json:"alias"`
	PKey   string `json:"pKey"`
	Secret string `json:"secret"`
}

type UserConfig struct {
	ClientId   string `json:"clientId"`
	Identities []IdentityConfig
}

func GetUserConfig(config DssBaseConfig, configDir string) (UserConfig, error) {
	var uc UserConfig
	if err := checkDir(configDir); err != nil {
		if err = os.Mkdir(configDir, 0o777); err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
	}
	ucf := filepath.Join(configDir, "clientConfig")
	bs, err := os.ReadFile(ucf)
	if err != nil {
		idc, err := GenIdentity("")
		if err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
		uc := UserConfig{ClientId: uuid.New().String(), Identities: []IdentityConfig{idc}}
		bs, err = json.Marshal(uc)
		if err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
		if err = os.WriteFile(ucf, bs, 0o666); err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
		return uc, nil
	}
	err = json.Unmarshal(bs, &uc)
	return uc, err
}

func GetHomeUserConfigPath(config DssBaseConfig) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("in GetHomeUserConfigPath: %v", err)
	}
	return filepath.Join(homeDir, ".cabri"), nil
}

func GetHomeUserConfig(config DssBaseConfig) (UserConfig, error) {
	configDir, err := GetHomeUserConfigPath(config)
	if err != nil {
		return UserConfig{}, fmt.Errorf("in GetHomeUserConfig: %v", err)
	}
	return GetUserConfig(config, configDir)
}
