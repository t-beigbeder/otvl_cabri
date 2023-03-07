package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type WebApiOptions struct {
	BaseOptions
	HasLog bool
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
	ure, err := GetUiRunEnv[WebApiOptions, *WebApiVars](ctx, false)
	if err != nil {
		return err
	}
	_ = ure
	for i := 0; i < len(args); i++ {
		dssType, addr, localPath, root, _ := CheckDssUrlMapping(args[i])
		var params cabridss.CreateNewParams
		if dssType == "obs" {
			params = cabridss.CreateNewParams{DssType: "obs", LocalPath: localPath, GetIndex: cabridss.GetPIndex}
		} else if dssType == "olf" {
			params = cabridss.CreateNewParams{DssType: "olf", Root: localPath, GetIndex: cabridss.GetPIndex}
		} else {
			panic("FIXME")
		}
		dss, err := cabridss.CreateOrNewDss(params)
		if err != nil {
			return err
		}
		config := cabridss.WebDssServerConfig{Dss: dss.(cabridss.HDss), HasLog: opts.HasLog, ShutdownCallback: func(err error) error {
			return err
		}}
		server, ok := vars.servers[addr]
		if ok {
			server.ConfigureApi(root, config, cabridss.WebDssServerConfigurator, nil)
		} else {
			vars.servers[addr], err = cabridss.NewWebDssServer(addr, root, config)
			if err != nil {
				return err
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
