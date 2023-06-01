package cabrisync

import (
	"context"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabrifsu"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
)

type BeVerboseFunc func(level int, line string)

// SyncOptions indicate how the Synchronize function should behave
type SyncOptions struct {
	Serial       bool // synchronization tasks will be serialized
	ReducerLimit int  // number of parallel I/O per DSS or zero if unlimited
	InDepth      bool // synchronize sub-namespaces content recursively
	Evaluate     bool // don't synchronize, just report work to be done
	BiDir        bool // bidirectional synchronization, the latest modified content wins,
	// if false synchronization is done from left to right
	KeepContent bool                           // don't remove content deleted from one side in other side
	NoCh        bool                           // don't evaluate checksum when not available, compare content's size and modification time
	NoACL       bool                           // don't evaluate ACL, use default ACL
	LeftMapACL  map[string][]cabridss.ACLEntry // left to right ACL user names mapping
	RightMapACL map[string][]cabridss.ACLEntry // right to left ACL user names mapping
	BeVerbose   BeVerboseFunc                  // callback for process verbosity
	RefDiag     *SyncRefDiag                   // a reference report for diagnosis
}

func makeDssWritable(dss cabridss.Dss, path string, options SyncOptions) error {
	fsy, ok := dss.(*cabridss.FsyDss)
	if !ok {
		return nil
	}
	if options.Evaluate || !options.InDepth {
		return nil
	}
	return cabrifsu.EnableWrite(fsy.GetAfs(), ufpath.Join(fsy.GetRoot(), path), true)
}

func doSynchronize(ctx context.Context, ldss cabridss.Dss, lpath string, rdss cabridss.Dss, rpath string, options SyncOptions) (report SyncReport) {
	report = SyncReport{}
	if _, err := ldss.GetMeta(cabridss.AppendSlashIf(lpath), false); err != nil {
		report.GErr = fmt.Errorf("left path \"%s\": no such entry (%v)", lpath, err)
		return
	}
	if _, err := rdss.GetMeta(cabridss.AppendSlashIf(rpath), false); err != nil {
		report.GErr = fmt.Errorf("right path \"%s\": no such entry (%v)", rpath, err)
		return
	}
	if err := makeDssWritable(rdss, rpath, options); err != nil {
		report.GErr = fmt.Errorf("makeDssWritable: %v", err)
	}
	if options.BiDir {
		if err := makeDssWritable(ldss, lpath, options); err != nil {
			report.GErr = fmt.Errorf("makeDssWritable: %v", err)
		}
	}
	syc := syncCtx{
		options: options,
		left:    sideCtx{options: options, dss: ldss, root: lpath, isNs: true, exist: true},
		right:   sideCtx{options: options, dss: rdss, isRight: true, root: rpath, isNs: true, exist: true},
	}
	report.Entries = syncNs(ctx, &syc)
	return
}

type SyncArgs struct {
	LDss  cabridss.Dss
	LPath string
	RDss  cabridss.Dss
	RPath string
	SOpts SyncOptions
}

func PlizedSynchronize(ctx context.Context, iInput interface{}) interface{} {
	is := iInput.(SyncArgs)
	sr := doSynchronize(ctx, is.LDss, is.LPath, is.RDss, is.RPath, is.SOpts)
	return sr
}

// Synchronize synchronizes namespaces and their contents between two DSS
// ctx is the parent context, possibly nil
// ldss is the left-side DSS
// lpath is the left-side namespace path
// rdss is the right-side DSS
// rpath is the right-side namespace path
// options are the options for the synchronization
// returns a report on the synchronization
func Synchronize(ctx context.Context, ldss cabridss.Dss, lpath string, rdss cabridss.Dss, rpath string, options SyncOptions) (report SyncReport) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = plumber.ContextWithConfig(ctx, cabridss.CabriPlumberDefaultConfig(options.Serial, options.ReducerLimit))
	iOutputs := plumber.LaunchAndWait(ctx,
		[]string{"Synchronized"},
		[]plumber.Launchable{PlizedSynchronize},
		[]interface{}{SyncArgs{LDss: ldss, LPath: lpath, RDss: rdss, RPath: rpath, SOpts: options}},
	)
	outputs := plumber.Retype[SyncReport](iOutputs)
	report = outputs[0]
	return
}
