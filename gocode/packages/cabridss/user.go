package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

type UserConfig struct {
	ClientId   string `json:"clientId"`
	Identities []IdentityConfig
}

func writeUserConfig(configDir string, uc UserConfig) error {
	bs, err := json.Marshal(uc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "clientConfig"), bs, 0o666)
}

func GetUserConfig(config DssBaseConfig, configDir string) (UserConfig, error) {
	var uc UserConfig
	if err := checkDir(configDir); err != nil {
		if err = os.Mkdir(configDir, 0o777); err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
	}
	bs, err := os.ReadFile(filepath.Join(configDir, "clientConfig"))
	if err != nil {
		idc, err := GenIdentity("")
		if err != nil {
			return uc, fmt.Errorf("in GetUserConfig: %w", err)
		}
		uc := UserConfig{ClientId: uuid.New().String(), Identities: []IdentityConfig{idc}}
		if err := writeUserConfig(configDir, uc); err != nil {
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

func UserConfigPutIdentity(config DssBaseConfig, configDir string, identity IdentityConfig) error {
	uc, err := GetUserConfig(config, configDir)
	if err != nil {
		return fmt.Errorf("in UserConfigPutIdentity %w", err)
	}
	over := false
	for i, cid := range uc.Identities {
		if cid.Alias == identity.Alias {
			uc.Identities[i] = identity
			over = true
			break
		}
	}
	if !over {
		uc.Identities = append(uc.Identities, identity)
	}
	return writeUserConfig(configDir, uc)
}

func IdPkeys(uc UserConfig) (users []string) {
	for _, id := range uc.Identities {
		users = append(users, id.PKey)
	}
	return
}

func IdSecrets(uc UserConfig) (secrets []string) {
	for _, id := range uc.Identities {
		secrets = append(secrets, id.Secret)
	}
	return
}
