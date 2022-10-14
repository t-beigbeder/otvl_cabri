package cabriui

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/joule"
)

type SampleOptions struct {
	BaseOptions
	FlagSample bool
}

type SampleVars struct {
	baseVars
	var1 string
	var2 string
}

func outer(ctx context.Context) {
	getUnitOfWork[SampleOptions, SampleVars](ctx).UiStrOut("from outer\n")
}

func SampleStartup(cr *joule.CLIRunner[SampleOptions]) error {
	_ = cr.AddUow("sample",
		func(ctx context.Context, work joule.UnitOfWork, i interface{}) (interface{}, error) {
			(*uiCtxFrom[SampleOptions, SampleVars](ctx)).vars = SampleVars{baseVars: baseVars{uow: work}, var1: "31", var2: "42"}
			work.UiStrOut(fmt.Sprintf("A message from work sample on stdout o %v a %v\n", cr.Opts, cr.Args))
			outer(ctx)
			return nil, nil
		})
	return nil
}

func SampleShutdown(cr *joule.CLIRunner[SampleOptions]) error {
	return nil
}
