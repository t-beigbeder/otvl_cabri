package cabriui

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type ConfigOptions struct {
	BaseOptions
	Encrypt bool
	Decrypt bool
	Dump    bool
	Gen     bool
	Get     bool
	Put     bool
	Remove  bool
}

type ConfigVars struct {
	baseVars
}

func ConfigStartup(cr *joule.CLIRunner[ConfigOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[ConfigOptions, *ConfigVars](ctx)).vars = &ConfigVars{baseVars: baseVars{uow: work}}
			return nil, configRun(ctx)
		})
	return nil
}

func ConfigShutdown(cr *joule.CLIRunner[ConfigOptions]) error {
	return cr.GetUow("command").GetError()
}

func configCtx(ctx context.Context) *uiContext[ConfigOptions, *ConfigVars] {
	return uiCtxFrom[ConfigOptions, *ConfigVars](ctx)
}

func configVars(ctx context.Context) *ConfigVars { return (*configCtx(ctx)).vars }

func configOpts(ctx context.Context) ConfigOptions { return (*configCtx(ctx)).opts }

func configUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[ConfigOptions, *ConfigVars](ctx)
}

func configOut(ctx context.Context, s string) { configUow(ctx).UiStrOut(s) }

func configErr(ctx context.Context, s string) { configUow(ctx).UiStrErr(s) }

func configEncrypt(ctx context.Context) error {
	mp, err := MasterPassword(configUow(ctx), configOpts(ctx).BaseOptions, 2)
	if err != nil {
		return err
	}
	cd, err := ConfigDir(configOpts(ctx).BaseOptions)
	if err != nil {
		return err
	}
	return cabridss.EncryptUserConfig(cabridss.DssBaseConfig{ConfigPassword: mp}, cd)
}

func configDecrypt(ctx context.Context) error {
	mp, err := MasterPassword(configUow(ctx), configOpts(ctx).BaseOptions, 1)
	if err != nil {
		return err
	}
	cd, err := ConfigDir(configOpts(ctx).BaseOptions)
	if err != nil {
		return err
	}
	return cabridss.DecryptUserConfig(cabridss.DssBaseConfig{ConfigPassword: mp}, cd)
}

func configDump(ctx context.Context) error {
	ure, err := GetUiRunEnv[ConfigOptions, *ConfigVars](ctx, false)
	if err != nil {
		return err
	}
	bs, err := json.MarshalIndent(ure.UserConfig, "", "  ")
	if err != nil {
		return err
	}
	configOut(ctx, string(bs)+"\n")
	return nil
}

func configGen(ctx context.Context) error {
	ure, err := GetUiRunEnv[ConfigOptions, *ConfigVars](ctx, false)
	if err != nil {
		return err
	}
	for _, alias := range configCtx(ctx).args {
		ic, err := cabridss.GenIdentity(alias)
		if err != nil {
			return err
		}
		ure.UserConfig.PutIdentity(ic)
	}
	return cabridss.SaveUserConfig(cabridss.DssBaseConfig{ConfigPassword: ure.MasterPassword}, ure.ConfigDir, ure.UserConfig)
}

func configGet(ctx context.Context) error {
	ure, err := GetUiRunEnv[ConfigOptions, *ConfigVars](ctx, false)
	if err != nil {
		return err
	}
	for _, alias := range configCtx(ctx).args {
		ic := ure.UserConfig.GetIdentity(alias)
		if ic.PKey == "" {
			return fmt.Errorf("identity for alias %s not found", alias)
		}
		configOut(ctx, fmt.Sprintf("PKey: %s\nSecret: %s\n", ic.PKey, ic.Secret))
	}
	return nil
}

func configPut(ctx context.Context) error {
	ure, err := GetUiRunEnv[ConfigOptions, *ConfigVars](ctx, false)
	if err != nil {
		return err
	}
	args := configCtx(ctx).args
	if len(args) < 2 {
		return fmt.Errorf("<alias> <pkey> not provided")
	}
	secret := ""
	if len(args) == 3 {
		secret = args[2]
		em, err := cabridss.EncryptMsg(secret, args[1])
		if err != nil {
			return err
		}
		dm, err := cabridss.DecryptMsg(em, secret)
		if err != nil {
			return err
		}
		if dm != secret {
			return fmt.Errorf("encryption error")
		}
	}

	ure.UserConfig.PutIdentity(cabridss.IdentityConfig{Alias: args[0], PKey: args[1], Secret: secret})
	return cabridss.SaveUserConfig(cabridss.DssBaseConfig{ConfigPassword: ure.MasterPassword}, ure.ConfigDir, ure.UserConfig)
}

func configRemove(ctx context.Context) error {
	ure, err := GetUiRunEnv[ConfigOptions, *ConfigVars](ctx, false)
	if err != nil {
		return err
	}
	args := configCtx(ctx).args
	for _, alias := range args {
		ic := ure.UserConfig.GetIdentity(alias)
		if ic.PKey == "" {
			return fmt.Errorf("identity for alias %s not found", alias)
		}
	}
	var newIds []cabridss.IdentityConfig
	for _, id := range ure.UserConfig.Identities {
		found := false
		for _, alias := range args {
			if id.Alias == alias {
				found = true
				break
			}
		}
		if !found {
			newIds = append(newIds, id)
		}
		ure.UserConfig.Identities = newIds
	}
	return cabridss.SaveUserConfig(cabridss.DssBaseConfig{ConfigPassword: ure.MasterPassword}, ure.ConfigDir, ure.UserConfig)
}

func configRun(ctx context.Context) error {
	opts := configOpts(ctx)
	err := fmt.Errorf("at least one operation option must be given with the config command")
	if opts.Encrypt {
		err = configEncrypt(ctx)
	}
	if opts.Decrypt {
		err = configDecrypt(ctx)
	}
	if opts.Dump {
		err = configDump(ctx)
	}
	if opts.Gen {
		err = configGen(ctx)
	}
	if opts.Get {
		err = configGet(ctx)
	}
	if opts.Put {
		err = configPut(ctx)
	}
	if opts.Remove {
		err = configRemove(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}
