package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"os"
	"strings"
)

type UiRunEnv struct {
	Encrypted       bool
	MasterPassword  string
	ConfigDir       string
	UserConfig      cabridss.UserConfig
	Users           []string
	UiACL           []cabridss.ACLEntry
	DefaultSyncUser string // only used for synchro: OS uid for fsy DSS, empty otherwise
}

func (ure UiRunEnv) GetACL() []cabridss.ACLEntry {
	if !ure.Encrypted {
		return ure.UiACL
	}
	var acl []cabridss.ACLEntry
	for _, uac := range ure.UiACL {
		if idc := ure.UserConfig.GetIdentity(uac.User); idc.PKey != "" {
			acl = append(acl, cabridss.ACLEntry{User: idc.PKey, Rights: uac.Rights})
		} else {
			acl = append(acl, uac)
		}
	}
	return acl
}

func (ure UiRunEnv) ACLOrDefault() ([]cabridss.ACLEntry, error) {
	if len(ure.UiACL) > 0 {
		return ure.GetACL(), nil
	}
	dr := cabridss.Rights{Read: true, Write: true}
	if !ure.Encrypted {
		return []cabridss.ACLEntry{{Rights: dr}}, nil
	}
	if idc := ure.UserConfig.GetIdentity(""); idc.PKey != "" {
		return []cabridss.ACLEntry{{User: idc.PKey, Rights: dr}}, nil
	}
	return nil, fmt.Errorf("in UiRunEnv.ACLOrDefault: no default public key")
}

func MasterPassword(uow joule.UnitOfWork, opts BaseOptions, askNumber int) (string, error) {
	if opts.PassFile != "" {
		bs, err := os.ReadFile(opts.PassFile)
		if err != nil {
			return "", err
		}
		if bs[len(bs)-1] == '\n' {
			bs = bs[:len(bs)-1]
		}
		return string(bs), nil
	}
	if askNumber > 0 || opts.Password {
		passwd1 := uow.UiSecret("please enter the master password: ")
		if askNumber > 1 {
			passwd2 := uow.UiSecret("please enter the master password again: ")
			if passwd1 != passwd2 || passwd1 == "" {
				return "", fmt.Errorf("passwords differ or are empty")
			}
		}
		return passwd1, nil
	}
	return "", nil
}

func ConfigDir(opts BaseOptions) (string, error) {
	cd := opts.ConfigDir
	var err error
	if cd == "" {
		cd, err = cabridss.GetHomeConfigDir(cabridss.DssBaseConfig{})
		if err != nil {
			return "", err
		}
	}
	fi, err := os.Stat(cd)
	if err != nil {
		if opts.ConfigDir == "" {
			if err = os.Mkdir(cd, 0o777); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a directory", cd)
	}
	return cd, nil
}

func GetUiRunEnv[OT BaseOptionsEr, VT baseVarsEr](ctx context.Context, encrypted bool) (ure UiRunEnv, err error) {
	uow := getUnitOfWork[OT, VT](ctx)
	uictx := uiCtxFrom[OT, VT](ctx)
	bo := uictx.opts.getBaseOptions()
	if ure.MasterPassword, err = MasterPassword(uow, bo, 0); err != nil {
		return
	}
	if ure.ConfigDir, err = ConfigDir(bo); err != nil {
		return
	}
	if ure.UserConfig, err = cabridss.GetUserConfig(cabridss.DssBaseConfig{ConfigPassword: ure.MasterPassword}, ure.ConfigDir); err != nil {
		return
	}
	if ure.UiACL, err = CheckUiACL(bo.ACL); err != nil {
		return
	}
	ure.Users = bo.Users
	ure.Encrypted = encrypted
	if _, err = ure.ACLOrDefault(); err != nil {
		return
	}
	return
}

type NewHDssArgs struct {
	Lasttime int64
	DssIx    int
	ObsIx    int
}

func NewHDss[OT BaseOptionsEr, VT baseVarsEr](
	ctx context.Context, setCfgFunc func(bc cabridss.DssBaseConfig), nhArgs NewHDssArgs,
) (cabridss.HDss, error) {
	uictx := uiCtxFrom[OT, VT](ctx)
	bo := uictx.opts.getBaseOptions()
	ucArgs := uictx.args
	var (
		dssType, root string
		err           error
	)
	dssType, root, _, err = CheckDssPath(ucArgs[nhArgs.DssIx])
	if err != nil {
		dssType, root, err = CheckDssSpec(ucArgs[nhArgs.DssIx])
	}
	ure, err := GetUiRunEnv[OT, VT](ctx, dssType[0] == 'x')
	if err != nil {
		return nil, err
	}
	var dss cabridss.HDss
	if dssType == "olf" {
		oc, err := GetOlfConfig(bo, nhArgs.ObsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, err
		}
		if setCfgFunc != nil {
			setCfgFunc(oc.DssBaseConfig)
		}
		if dss, err = cabridss.NewOlfDss(oc, nhArgs.Lasttime, nil); err != nil {
			return nil, err
		}
	} else if dssType == "xolf" {
		if dss, err = NewXolfDss(bo, nhArgs.ObsIx, nhArgs.Lasttime, root, ure.MasterPassword, ure.Users); err != nil {
			return nil, err
		}
	} else if dssType == "obs" {
		oc, err := GetObsConfig(bo, nhArgs.ObsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, err
		}
		if setCfgFunc != nil {
			setCfgFunc(oc.DssBaseConfig)
		}
		if dss, err = cabridss.NewObsDss(oc, nhArgs.Lasttime, nil); err != nil {
			return nil, err
		}
	} else if dssType == "xobs" {
		if dss, err = NewXobsDss(bo, nhArgs.ObsIx, nhArgs.Lasttime, root, ure.MasterPassword, ure.Users); err != nil {
			return nil, err
		}
	} else if dssType == "smf" {
		sc, err := GetSmfConfig(bo, nhArgs.ObsIx, root, ure.MasterPassword)
		if err != nil {
			return nil, err
		}
		if setCfgFunc != nil {
			setCfgFunc(sc.DssBaseConfig)
		}
		if dss, err = cabridss.NewObsDss(sc, nhArgs.Lasttime, nil); err != nil {
			return nil, err
		}
	} else if dssType == "webapi+http" {
		frags := strings.Split(root[2:], "/")
		wc, err := GetWebConfig(bo, nhArgs.ObsIx, frags[0], frags[1], ure.MasterPassword)
		if err != nil {
			return nil, err
		}
		if dss, err = cabridss.NewWebDss(wc, nhArgs.Lasttime, ure.Users); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("DSS type %s is not (yet) supported", dssType)
	}
	return dss, nil
}

func NewXolfDss(opts BaseOptions, index int, lasttime int64, root, mp string, aclusers []string) (cabridss.HDss, error) {
	oc, err := GetOlfConfig(opts, index, root, mp)
	if err != nil {
		return nil, err
	}
	bc, err := GetBaseConfig(opts, index, root, root, mp)
	if err != nil {
		return nil, err
	}
	if bc.GetIndex == nil {
		oc.GetIndex = cabridss.GetPIndex
	}
	dss, err := cabridss.NewEDss(
		cabridss.EDssConfig{
			WebDssConfig: cabridss.WebDssConfig{
				DssBaseConfig: cabridss.DssBaseConfig{
					LibApi:         true,
					ConfigDir:      oc.ConfigDir,
					ConfigPassword: mp,
				},
				LibApiDssConfig: cabridss.LibApiDssConfig{
					IsOlf:  true,
					OlfCfg: oc,
				},
			},
		},
		lasttime, aclusers)
	return dss, err
}

func NewXobsDss(opts BaseOptions, index int, lasttime int64, root, mp string, aclusers []string) (cabridss.HDss, error) {
	oc, err := GetObsConfig(opts, index, root, mp)
	if err != nil {
		return nil, err
	}
	bc, err := GetBaseConfig(opts, index, root, root, mp)
	if err != nil {
		return nil, err
	}
	if bc.GetIndex == nil {
		oc.GetIndex = cabridss.GetPIndex
	}
	dss, err := cabridss.NewEDss(
		cabridss.EDssConfig{
			WebDssConfig: cabridss.WebDssConfig{
				DssBaseConfig: cabridss.DssBaseConfig{
					LibApi:         true,
					ConfigDir:      oc.ConfigDir,
					ConfigPassword: mp,
				},
				LibApiDssConfig: cabridss.LibApiDssConfig{
					IsObs:  true,
					ObsCfg: oc,
				},
			},
		},
		lasttime, aclusers)
	return dss, err
}
