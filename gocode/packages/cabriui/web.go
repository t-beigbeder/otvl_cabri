package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type WebApiOptions struct {
	BaseOptions
	HasLog        bool
	IsRest        bool
	TlsKey        string // certificate key file on https server
	LastTime      string
	TlsClientCert string // untrusted CA on https client
}

func (wos WebApiOptions) getLastTime() (lastTime int64) {
	if wos.LastTime != "" {
		lastTime, _ = CheckTimeStamp(wos.LastTime)
	}
	return
}

type WebApiVars struct {
	baseVars
	servers map[string]cabridss.WebServer
}

func WebApiStartup(cr *joule.CLIRunner[WebApiOptions]) error {
	_ = cr.AddUow("command",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[WebApiOptions, *WebApiVars](ctx)).vars =
				&WebApiVars{baseVars: baseVars{uow: work}, servers: map[string]cabridss.WebServer{}}
			return nil, webApi(ctx, cr.Args)
		})
	return nil
}

func WebApiShutdown(cr *joule.CLIRunner[WebApiOptions]) error {
	return cr.GetUow("command").GetError()
}

func webApiCtx(ctx context.Context) *uiContext[WebApiOptions, *WebApiVars] {
	return uiCtxFrom[WebApiOptions, *WebApiVars](ctx)
}

func webApiVars(ctx context.Context) *WebApiVars { return (*webApiCtx(ctx)).vars }

func webApiOpts(ctx context.Context) WebApiOptions { return (*webApiCtx(ctx)).opts }

func webApiUow(ctx context.Context) joule.UnitOfWork {
	return getUnitOfWork[WebApiOptions, *WebApiVars](ctx)
}

func webApiOut(ctx context.Context, s string) { webApiUow(ctx).UiStrOut(s) }

func webApiErr(ctx context.Context, s string) { webApiUow(ctx).UiStrErr(s) }

func webApi(ctx context.Context, args []string) error {
	opts := webApiOpts(ctx)
	vars := webApiVars(ctx)
	ure, err := GetUiRunEnv[WebApiOptions, *WebApiVars](ctx, false, false)
	if err != nil {
		return err
	}
	_ = ure
	obsIx := 0
	for i := 0; i < len(args); i++ {
		dssType, addr, localPath, root, isTls, _ := CheckDssUrlMapping(args[i])
		if isTls && (opts.TlsCert == "" || opts.TlsKey == "") {
			return fmt.Errorf("mapping %s requires certificate and key files", args[i])
		}
		var dss cabridss.Dss
		var err error
		if !opts.IsRest {
			var params cabridss.CreateNewParams
			dssSubType := dssType
			if dssType[0] == 'x' {
				dssSubType = dssType[1:]
			}
			if dssSubType == "obs" || dssSubType == "smf" {
				params = cabridss.CreateNewParams{DssType: dssType, LocalPath: localPath, ConfigPassword: ure.MasterPassword}
			} else if dssSubType == "olf" {
				params = cabridss.CreateNewParams{DssType: "olf", Root: localPath, ConfigPassword: ure.MasterPassword}
			}
			params.Encrypted = dssType[0] == 'x'
			dss, err = cabridss.CreateOrNewDss(params)
		} else {
			dss, err = NewHDss[WebApiOptions, *WebApiVars](ctx, nil, NewHDssArgs{DssIx: i, ObsIx: obsIx, Lasttime: webApiOpts(ctx).getLastTime(), IsMapping: true})
		}
		if err != nil {
			return err
		}
		if dss.(cabridss.HDss).GetIndex() == nil || !dss.(cabridss.HDss).GetIndex().IsPersistent() {
			return fmt.Errorf("DSS for url %s is not persistent", args[i])
		}
		config := cabridss.WebDssServerConfig{
			WebServerConfig: cabridss.WebServerConfig{
				Addr:              addr,
				HasLog:            opts.HasLog,
				IsTls:             isTls,
				TlsCert:           opts.TlsCert,
				TlsKey:            opts.TlsKey,
				TlsNoCheck:        opts.TlsNoCheck,
				BasicAuthUser:     ure.BasicAuthUser,
				BasicAuthPassword: ure.BasicAuthPassword,
			},
			Dss: dss.(cabridss.HDss),
		}
		if opts.IsRest {
			config.UserConfig = ure.UserConfig
		}
		server, ok := vars.servers[addr]
		if !ok {
			var err error
			if opts.IsRest {
				vars.servers[addr], err = cabridss.NewRestServer(root, config)
			} else {
				vars.servers[addr], err = cabridss.NewWebDssServer(root, config)
			}
			if err != nil {
				dss.Close()
				return err
			}
		} else {
			if opts.IsRest {
				server.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
					return customConfigs[root].(cabridss.WebDssServerConfig).Dss.Close()
				}, cabridss.RestServerConfigurator)
			} else {
				server.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
					return customConfigs[root].(cabridss.WebDssServerConfig).Dss.Close()
				}, cabridss.WebDssServerConfigurator)
			}
		}
	}
	<-ctx.Done()
	for addr, server := range vars.servers {
		webApiErr(ctx, fmt.Sprintf("server at %s shutting down\n", addr))
		if err := server.Shutdown(); err != nil {
			webApiErr(ctx, fmt.Sprintf("server at %s shutdown failed with error %v\n", addr, err))
		}
	}
	return nil
}
