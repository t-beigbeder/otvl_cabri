package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type BaseOptions struct {
	ConfigDir     string
	IdentityAlias string
	Password      bool
	PassFile      string
	Serial        bool
	IndexImplems  []string
	ObsRegions    []string
	ObsEndpoints  []string
	ObsContainers []string
	ObsAccessKeys []string
	ObsSecretKeys []string
}

func (bos BaseOptions) getBaseOptions() BaseOptions {
	return bos
}

type BaseOptionsEr interface {
	getBaseOptions() BaseOptions
}

type baseVars struct {
	uow joule.UnitOfWork
}

func (bvs baseVars) getUnitOfWork() joule.UnitOfWork {
	return bvs.uow
}

type baseVarsEr interface {
	getUnitOfWork() joule.UnitOfWork
}

type cabriUiKey int

const uiCtxKey cabriUiKey = 1

type uiContext[OT BaseOptionsEr, VT baseVarsEr] struct {
	opts OT
	args []string
	vars VT
}

func CLIRun[OT BaseOptionsEr, VT baseVarsEr](
	cliIn io.Reader, cliOut io.Writer, cliErr io.Writer,
	opts OT, args []string,
	startup func(cr *joule.CLIRunner[OT]) error,
	shutdown func(cr *joule.CLIRunner[OT]) error,
) error {

	cliStartup := func(cr *joule.CLIRunner[OT]) error {
		pCtx := plumber.ContextWithConfig(*cr.Ctx, cabridss.CabriPlumberDefaultConfig(opts.getBaseOptions().Serial))
		ctx := context.WithValue(pCtx, uiCtxKey, &uiContext[OT, VT]{opts: opts, args: args})
		cr.Ctx = &ctx
		return startup(cr)
	}

	cliShutdown := func(cr *joule.CLIRunner[OT]) error {
		return shutdown(cr)
	}

	cr := joule.NewCLIRunner(opts, args, cliIn, cliOut, cliErr, cliStartup, cliShutdown)
	return cr.Run()
}

func uiCtxFrom[OT BaseOptionsEr, VT baseVarsEr](ctx context.Context) *uiContext[OT, VT] {
	v := ctx.Value(uiCtxKey)
	_ = v
	uiCtx, _ := ctx.Value(uiCtxKey).(*uiContext[OT, VT])
	return uiCtx
}

func getUnitOfWork[OT BaseOptionsEr, VT baseVarsEr](ctx context.Context) joule.UnitOfWork {
	return (*uiCtxFrom[OT, VT](ctx)).vars.getUnitOfWork()
}

func CheckDssSpec(dssSpec string) (dssType, root string, err error) {
	frags := strings.Split(dssSpec, ":")
	if len(frags) != 2 {
		err = fmt.Errorf("DSS specification %s is invalid", dssSpec)
		return
	}
	if frags[0] != "fsy" && frags[0] != "olf" && frags[0] != "xolf" && frags[0] != "obs" && frags[0] != "smf" {
		err = fmt.Errorf("DSS type %s is not (yet) supported", frags[0])
		return
	}
	dssType = frags[0]
	root = frags[1]
	return
}

func CheckDssPath(dssPath string) (dssType, root, npath string, err error) {
	frags := strings.Split(dssPath, ":")
	if len(frags) < 2 {
		err = fmt.Errorf("DSS path %s is invalid", dssPath)
		return
	}
	if len(frags) > 2 {
		frags[1] = strings.Join(frags[1:], ":")
	}
	if frags[0] != "fsy" && frags[0] != "olf" && frags[0] != "xolf" && frags[0] != "obs" && frags[0] != "smf" && frags[0] != "webapi+http" {
		err = fmt.Errorf("DSS type %s is not (yet) supported", frags[0])
		return
	}
	dssType = frags[0]
	if dssType == "webapi+http" && (!strings.HasPrefix(frags[1], "//") || len(strings.Split(frags[1][2:], "/")) < 2) {
		err = fmt.Errorf("DSS type %s requires //host[:port]/[path] url syntax (in %s)", frags[0], frags[1])
		return
	}
	rnPath := strings.Split(frags[1], "@")
	if len(rnPath) != 2 {
		err = fmt.Errorf("DSS root/path %s is invalid", rnPath)
		return
	}
	root = rnPath[0]
	npath = rnPath[1]
	return
}

func CheckDssUrlMapping(dum string) (dssType, addr, localPath, root string, err error) {
	frags := strings.Split(dum, "://")
	if len(frags) != 2 || !strings.HasSuffix(frags[0], "+http") {
		err = fmt.Errorf("DSS URL mapping %s is invalid", dum)
		return
	}
	dssType = frags[0][:len(frags[0])-5]
	if dssType != "fsy" && dssType != "olf" && dssType != "obs" && dssType != "smf" {
		err = fmt.Errorf("DSS type %s is not (yet) supported", dssType)
		return
	}
	rFrags := strings.Split(frags[1], "/")
	if len(rFrags) < 2 {
		err = fmt.Errorf("DSS URL mapping %s is invalid", dum)
		return
	}
	addr = rFrags[0]
	r2Frags := strings.Split(frags[1][len(addr):], "@")
	if len(r2Frags) != 2 {
		err = fmt.Errorf("DSS URL mapping %s is invalid", dum)
		return
	}
	localPath = r2Frags[0]
	root = r2Frags[1]
	return
}

func CheckTimeStamp(value string) (unix int64, err error) {
	if value == "" {
		return
	}
	var ts time.Time
	if ts, err = time.Parse(time.RFC3339, value); err == nil {
		unix = ts.Unix()
		return
	}
	if unix, err = strconv.ParseInt(value, 10, 64); err == nil {
		return
	}
	err = fmt.Errorf("timestamp %s must be either RFC3339 or a unix time integer", value)
	return
}

func GetBaseConfig(opts BaseOptions, index int, root, localPath, mp string) (cabridss.DssBaseConfig, error) {
	cd, err := ConfigDir(opts)
	if err != nil {
		return cabridss.DssBaseConfig{}, nil
	}
	dbc := cabridss.DssBaseConfig{
		ConfigDir:      cd,
		ConfigPassword: mp,
		LocalPath:      localPath}

	if len(opts.IndexImplems) > index {
		if opts.IndexImplems[index] == "no" {
			dbc.GetIndex = func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
				return cabridss.NewNIndex(), nil
			}
		} else if opts.IndexImplems[index] == "memory" {
			dbc.GetIndex = func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
				return cabridss.NewMIndex(), nil
			}
		} else if opts.IndexImplems[index] == "bdb" {
			dbc.GetIndex = cabridss.GetPIndex
		} else {
			return cabridss.DssBaseConfig{}, fmt.Errorf("index implementation #%d is unknown %s (no, memory, bdb)", index+1, opts.IndexImplems[index])
		}
	}
	return dbc, nil
}

func GetOlfConfig(opts BaseOptions, index int, root, mp string) (cabridss.OlfConfig, error) {
	bc, err := GetBaseConfig(opts, index, root, root, mp)
	if err != nil {
		return cabridss.OlfConfig{}, err
	}
	if bc.GetIndex == nil {
		bc.GetIndex = func(config cabridss.DssBaseConfig, _ string) (cabridss.Index, error) {
			return cabridss.NewMIndex(), nil
		}
	}
	return cabridss.OlfConfig{DssBaseConfig: bc, Root: root}, nil
}

func GetObsConfig(opts BaseOptions, index int, root, mp string) (cabridss.ObsConfig, error) {
	var region, endpoint, container, accessKey, secretKey string
	if len(opts.ObsRegions) > index {
		region = opts.ObsRegions[index]
	}
	if len(opts.ObsEndpoints) > index {
		endpoint = opts.ObsEndpoints[index]
	}
	if len(opts.ObsContainers) > index {
		container = opts.ObsContainers[index]
	}
	if len(opts.ObsAccessKeys) > index {
		accessKey = opts.ObsAccessKeys[index]
	}
	if len(opts.ObsSecretKeys) > index {
		secretKey = opts.ObsSecretKeys[index]
	}
	bc, err := GetBaseConfig(opts, index, root, root, mp)
	if err != nil {
		return cabridss.ObsConfig{}, err
	}
	if bc.GetIndex == nil {
		bc.GetIndex = cabridss.GetPIndex
	}
	return cabridss.ObsConfig{
		DssBaseConfig: bc,
		Endpoint:      endpoint,
		Region:        region,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		Container:     container,
	}, nil
}

func GetSmfConfig(opts BaseOptions, index int, root, mp string) (cabridss.ObsConfig, error) {
	config, err := GetObsConfig(opts, index, root, mp)
	if err != nil {
		return cabridss.ObsConfig{}, err
	}
	config.GetS3Session = func() cabridss.IS3Session {
		return cabridss.NewS3sMockFs(root, nil)
	}
	return config, nil
}

func GetWebConfig(opts BaseOptions, index int, addr, root, mp string) (cabridss.WebDssConfig, error) {
	bc, err := GetBaseConfig(opts, index, "", "", mp)
	if err != nil {
		return cabridss.WebDssConfig{}, err
	}
	var port string
	frags := strings.Split(addr, ":")
	host := frags[0]
	if len(frags) > 1 {
		port = frags[1]
	}
	bc.WebHost = host
	bc.WebPort = port
	bc.WebRoot = root
	return cabridss.WebDssConfig{DssBaseConfig: bc}, nil
}

func MutualExcludeFlags(names []string, flags ...bool) error {
	for i, bname := range names {
		bval := flags[i]
		for j := i + 1; j < len(names); j++ {
			oval := flags[j]
			if bval && oval {
				return fmt.Errorf(fmt.Sprintf("flags %s and %s cannot be used at the same time", bname, names[j]))
			}
		}
	}
	return nil
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
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a directory", cd)
	}
	return cd, nil
}

func NewXolfDss(opts BaseOptions, index int, lasttime int64, root, mp string) (cabridss.HDss, error) {
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
		lasttime, nil)
	return dss, err
}
