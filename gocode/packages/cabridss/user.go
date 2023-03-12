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
	Internal   IdentityConfig
}

func (uc *UserConfig) PutIdentity(identity IdentityConfig) {
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
}

func (uc *UserConfig) GetIdentity(alias string) IdentityConfig {
	for _, cid := range uc.Identities {
		if cid.Alias == alias {
			return cid
		}
	}
	return IdentityConfig{}
}

func ccPath(configDir string) string {
	return filepath.Join(configDir, "clientConfig")
}

func writeUserConfig(config DssBaseConfig, configDir string, uc UserConfig, encrypt bool) error {
	bs, err := json.Marshal(uc)
	if err != nil {
		return err
	}
	if encrypt {
		if config.ConfigPassword == "" {
			return fmt.Errorf("in writeUserConfig: no password provided for configuration encryption")
		}
		if bs, err = EncryptMsgWithPass(string(bs), config.ConfigPassword); err != nil {
			return fmt.Errorf("in writeUserConfig: %w", err)
		}
	}
	return os.WriteFile(ccPath(configDir), bs, 0o666)
}

func getUserConfigInfo(config DssBaseConfig, configDir string) (UserConfig, bool, error) {
	var uc UserConfig
	if err := checkDir(configDir); err != nil {
		if err = os.Mkdir(configDir, 0o777); err != nil {
			return uc, false, err
		}
	}
	bs, err := os.ReadFile(ccPath(configDir))
	if err != nil {
		defIdc, err := GenIdentity("")
		if err != nil {
			return uc, false, err
		}
		internIdc, err := GenIdentity("__internal__")
		if err != nil {
			return uc, false, err
		}
		uc := UserConfig{ClientId: uuid.New().String(), Identities: []IdentityConfig{defIdc}, Internal: internIdc}
		if err := writeUserConfig(config, configDir, uc, false); err != nil {
			return uc, false, err
		}
		return uc, false, nil
	}
	encrypted := false
	if err = json.Unmarshal(bs, &uc); err != nil {
		if config.ConfigPassword == "" {
			return uc, false, ErrPasswordRequired
		}
		sbs, err2 := DecryptMsgWithPass(bs, config.ConfigPassword)
		if err2 != nil {
			return uc, false, err2
		}
		encrypted = true
		err = json.Unmarshal([]byte(sbs), &uc)
	}
	return uc, encrypted, err
}

func GetUserConfig(config DssBaseConfig, configDir string) (UserConfig, error) {
	uc, _, err := getUserConfigInfo(config, configDir)
	if err != nil {
		err = fmt.Errorf("in GetUserConfig: %w", err)
	}
	return uc, err
}

func SaveUserConfig(config DssBaseConfig, configDir string, uc UserConfig) error {
	_, enc, err := getUserConfigInfo(config, configDir)
	if err != nil {
		err = fmt.Errorf("in GetUserConfig: %w", err)
	}
	return writeUserConfig(config, configDir, uc, enc)
}

func GetHomeConfigDir(config DssBaseConfig) (string, error) {
	homeDir, err := OsUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("in GetHomeConfigDir: %v", err)
	}
	return filepath.Join(homeDir, ".cabri"), nil
}

func GetHomeUserConfig(config DssBaseConfig) (UserConfig, error) {
	configDir, err := GetHomeConfigDir(config)
	if err != nil {
		return UserConfig{}, fmt.Errorf("in GetHomeUserConfig: %v", err)
	}
	return GetUserConfig(config, configDir)
}

func UserConfigPutIdentity(config DssBaseConfig, configDir string, identity IdentityConfig) error {
	uc, enc, err := getUserConfigInfo(config, configDir)
	if err != nil {
		return fmt.Errorf("in UserConfigPutIdentity: %w", err)
	}
	uc.PutIdentity(identity)
	return writeUserConfig(config, configDir, uc, enc)
}

func CurrentUserConfig(config DssBaseConfig) (uc UserConfig, err error) {
	if config.ConfigDir == "" {
		uc, err = GetHomeUserConfig(config)
	} else {
		uc, err = GetUserConfig(config, config.ConfigDir)
	}
	return
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

func EncryptUserConfig(config DssBaseConfig, configDir string) error {
	if config.ConfigPassword == "" {
		return fmt.Errorf("in EncryptUserConfig: empty password")
	}
	if _, encrypted, err := getUserConfigInfo(config, configDir); err != nil || encrypted {
		if err != nil {
			return fmt.Errorf("in EncryptUserConfig: %w", err)
		}
		if encrypted {
			return fmt.Errorf("in EncryptUserConfig: already encrypted")
		}
	}
	return EncryptFileWithPass(ccPath(configDir), ccPath(configDir), config.ConfigPassword)
}

func DecryptUserConfig(config DssBaseConfig, configDir string) error {
	if _, encrypted, err := getUserConfigInfo(config, configDir); err != nil || !encrypted {
		if err != nil {
			return fmt.Errorf("in DecryptUserConfig: %w", err)
		}
		if !encrypted {
			return fmt.Errorf("in DecryptUserConfig: already cleartext")
		}
	}
	return DecryptFileWithPass(ccPath(configDir), ccPath(configDir), config.ConfigPassword)
}
