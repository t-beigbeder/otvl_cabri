package cabrisync

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
)

func plizedSyncEntries(ctx context.Context, ins interface{}) interface{} {
	var aggEntries []SyncReportEntry
	chsSyc := plumber.Retype[syncCtx](plumber.Untype[syncCtx](ins.([]syncCtx)))
	for _, entries := range plumber.Parallelize[syncCtx, []SyncReportEntry](
		ctx, "",
		func(ctx context.Context, syc syncCtx) []SyncReportEntry {
			if syc.left.isNs {
				return syncNs(ctx, &syc)
			} else {
				return syncContentOrSymLink(ctx, &syc)
			}
		},
		chsSyc...) {
		aggEntries = append(aggEntries, entries...)
	}
	return aggEntries
}

type lrMetaOut struct {
	isRight bool
	err     error
}

func plizedGetLRMeta(ctx context.Context, ins interface{}) interface{} {
	var outs []lrMetaOut
	sdcs := plumber.Retype[*sideCtx](plumber.Untype[*sideCtx](ins.([]*sideCtx)))
	for _, out := range plumber.Parallelize[*sideCtx, lrMetaOut](
		ctx, "",
		func(ctx context.Context, sdc *sideCtx) lrMetaOut {
			if sdc.isNs {
				return lrMetaOut{isRight: sdc.isRight, err: sdc.lsnsMeta()}
			} else {
				return lrMetaOut{isRight: sdc.isRight, err: sdc.getMeta()}
			}
		},
		sdcs...,
	) {
		outs = append(outs, out)
	}
	return outs
}

func lrMetaErrs(outs []interface{}) (lErr, rErr error) {
	if outs == nil || len(outs) != 1 || outs[0] == nil || len(outs[0].([]lrMetaOut)) != 2 {
		lErr = fmt.Errorf("GetLRMeta parallelization failed %v", outs)
		rErr = lErr
		return
	}
	if outs[0].([]lrMetaOut)[0].isRight {
		lErr = outs[0].([]lrMetaOut)[1].err
		rErr = outs[0].([]lrMetaOut)[0].err
	} else {
		lErr = outs[0].([]lrMetaOut)[0].err
		rErr = outs[0].([]lrMetaOut)[1].err
	}
	return
}

func syncNs(ctx context.Context, syc *syncCtx) []SyncReportEntry {
	syc.diagnose(">syncNs", false)
	rent := SyncReportEntry{IsNs: true, LPath: syc.left.fullPath(), RPath: syc.right.fullPath()}
	if syc.err != nil {
		rent.Err = syc.err
		syc.diagnose(fmt.Sprintf("<syncNs %v", syc.err), true)
		return []SyncReportEntry{rent}
	}
	iLrOuts := plumber.LaunchAndWait(ctx,
		[]string{"SyncGetMetas"},
		[]plumber.Launchable{plizedGetLRMeta},
		[]interface{}{[]*sideCtx{&syc.left, &syc.right}})
	leftErr, rightErr := lrMetaErrs(iLrOuts)
	if (leftErr != nil && !syc.options.BiDir) || (leftErr != nil && rightErr != nil) {
		syc.err = leftErr
		rent.Err = syc.err
		syc.diagnose(fmt.Sprintf("<syncNs %v", syc.err), true)
		return []SyncReportEntry{rent}
	}

	syc.eval(&rent)

	syc.evalNsMerge()
	if !syc.options.Evaluate {
		syc.mergeNsBefore(rent)
	}

	chsSyc := make([]syncCtx, 0)
	for _, pch := range syc.leftAndRight {
		isNs := pch[len(pch)-1] == '/'
		if isNs && !syc.options.InDepth {
			continue
		}
		chsSyc = append(chsSyc, syc.makeChild(pch))
	}
	iOuts := plumber.LaunchAndWait(ctx,
		[]string{"SyncEntries"},
		[]plumber.Launchable{plizedSyncEntries},
		[]interface{}{chsSyc})
	_ = iOuts

	entries := make([]SyncReportEntry, 0)
	for _, chEntries := range plumber.Retype[[]SyncReportEntry](iOuts) {
		for _, se := range chEntries {
			entries = append(entries, se)
		}
	}

	if !syc.options.Evaluate {
		syc.mergeNsAfter(rent)
	}
	if syc.options.RefDiag != nil {
		if syc.options.RefDiag.Left[rent.LPath] != rent {
			syc.diagnose("ns panic", false)
		}
	}
	entries = append(entries, rent)
	syc.diagnose("<syncNs", true)
	return entries
}

func syncContentOrSymLink(ctx context.Context, syc *syncCtx) []SyncReportEntry {
	syc.diagnose(">syncContentOrSymLink", false)
	rent := SyncReportEntry{IsNs: false, LPath: syc.left.fullPath(), RPath: syc.right.fullPath()}
	if syc.err != nil {
		rent.Err = syc.err
		syc.diagnose(fmt.Sprintf("<syncContentOrSymLink %v", syc.err), true)
		return []SyncReportEntry{rent}
	}
	iLrOuts := plumber.LaunchAndWait(ctx,
		[]string{"SyncGetMetas"},
		[]plumber.Launchable{plizedGetLRMeta},
		[]interface{}{[]*sideCtx{&syc.left, &syc.right}})
	leftErr, rightErr := lrMetaErrs(iLrOuts)
	if (leftErr != nil && !syc.options.BiDir) || (leftErr != nil && rightErr != nil) {
		syc.err = leftErr
		rent.Err = syc.err
		syc.diagnose(fmt.Sprintf("<syncContentOrSymLink %v", syc.err), true)
		return []SyncReportEntry{rent}
	}

	syc.eval(&rent)
	if syc.options.RefDiag != nil {
		if syc.options.RefDiag.Left[rent.LPath] != rent {
			syc.diagnose("content panic", false)
			_ = syc.left.getMeta()
			_ = syc.right.getMeta()
		}
	}
	if !syc.options.Evaluate {
		if syc.err == nil && (rent.Created || rent.Updated || rent.MUpdated) {
			if rent.isSymLink {
				syc.err = syc.crUpSymLink(rent.isRTL)
			} else {
				syc.err = syc.crUpContent(rent.isRTL)
			}
			if syc.err != nil {
				rent.Err = syc.err
			}
		}
	}
	syc.diagnose("<syncContentOrSymLink", true)
	return []SyncReportEntry{rent}
}
