package cabriui

import (
	"context"
	"errors"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"io"
	"strconv"
	"strings"
	"time"
)

type BaseOptions struct {
	ConfigDir     string
	Users         []string
	ACL           []string
	Password      bool
	PassFile      string
	Serial        bool
	IndexImplems  []string
	ObsRegions    []string
	ObsEndpoints  []string
	ObsContainers []string
	ObsAccessKeys []string
	ObsSecretKeys []string
	TlsCert       string // certificate file on https server or untrusted CA on https client
	TlsNoCheck    bool   // no check of certifcate by https client
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
	err := cr.Run()
	if errors.Is(err, cabridss.ErrPasswordRequired) {
		return err // nothing, just to know
	}
	return err
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
	var npath string
	if dssType, root, npath, err = CheckDssPath(dssSpec + "@"); err != nil || npath != "" {
		err = fmt.Errorf("DSS specification %s is invalid", dssSpec)
		return
	}
	return
}

var dssTypes map[string]int = map[string]int{
	"fsy": 1, "olf": 1, "xolf": 1, "obs": 1, "xobs": 1, "smf": 1, "xsmf": 1,
	"webapi+http": 1, "webapi+https": 1,
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
	if tt, ok := dssTypes[frags[0]]; !ok || tt != 1 {
		err = fmt.Errorf("DSS type %s is not (yet) supported", frags[0])
		return
	}
	dssType = frags[0]
	if (dssType == "webapi+http" || dssType == "webapi+https") && (!strings.HasPrefix(frags[1], "//") || len(strings.Split(frags[1][2:], "/")) < 2) {
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

func CheckUiACL(sacl []string) (acl []cabridss.ACLEntry, err error) {
	for _, sac := range sacl {
		sacsubs := strings.Split(sac, ":")
		if len(sacsubs) != 2 {
			return nil, fmt.Errorf("invalid ACL string %s, not <user:rights>", sac)
		}
		u, rights := sacsubs[0], sacsubs[1]
		ur := cabridss.Rights{}
		for _, char := range rights {
			if char == 'r' {
				ur.Read = true
			} else if char == 'w' {
				ur.Write = true
			} else if char == 'x' {
				ur.Execute = true
			} else {
				return nil, fmt.Errorf("invalid character %c for access right (not in 'rwx')", char)
			}
		}
		acl = append(acl, cabridss.ACLEntry{User: u, Rights: ur})
	}
	return
}

func CheckDssUrlMapping(dum string) (dssType, addr, localPath, root string, isTls bool, err error) {
	frags := strings.Split(dum, "://")
	if len(frags) != 2 || (!strings.HasSuffix(frags[0], "+http") && !strings.HasSuffix(frags[0], "+https")) {
		err = fmt.Errorf("DSS URL mapping %s is invalid", dum)
		return
	}
	dssType = frags[0][:strings.Index(frags[0], "+http")]
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
	isTls = strings.HasSuffix(frags[0], "+https")
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
	err = fmt.Errorf("timestamp %s must be either RFC3339 (eg 2020-08-13T11:56:41Z) or a unix time integer", value)
	return
}

func GetBaseConfig(opts BaseOptions, index int, root, localPath, mp string) (cabridss.DssBaseConfig, error) {
	cd, err := ConfigDir(opts)
	if err != nil {
		return cabridss.DssBaseConfig{}, err
	}
	dbc := cabridss.DssBaseConfig{
		ConfigDir:      cd,
		ConfigPassword: mp,
		LocalPath:      localPath,
	}

	if len(opts.IndexImplems) > index {
		dbc.XImpl = opts.IndexImplems[index]
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

func GetWebConfig(opts BaseOptions, index int, isTls bool, addr, root, mp string) (cabridss.WebDssConfig, error) {
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
	if isTls {
		bc.WebProtocol = "https"
		bc.TlsCert = opts.TlsCert
		bc.TlsNoCheck = opts.TlsNoCheck
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
