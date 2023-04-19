package cabrisync

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
)

type sideCtx struct {
	options     SyncOptions
	dss         cabridss.Dss
	isRight     bool
	root        string
	pPath       string
	isNs        bool
	path        string
	exist       bool     // meta exists
	created     bool     // actually created
	actualMtime int64    // actual mtime
	exCh        []string // existing children
	meta        cabridss.IMeta
}

type syncCtx struct {
	options      SyncOptions
	err          error
	left         sideCtx
	right        sideCtx
	leftAndRight []string // left and right possible children
	leftMg       []string // left merged (existing + added) children
	rightMg      []string // right merged (existing + added + keep removed) children
	rmRight      []string // right children removed
	leftRight    []string // right children left after remove
}

func (sdc *sideCtx) arrow() rune {
	if sdc.isRight {
		return '>'
	}
	return '<'
}

func (sdc *sideCtx) relPath() string {
	fp := ""
	if len(sdc.pPath) > 0 {
		if len(fp) > 0 {
			fp += "/" + sdc.pPath
		} else {
			fp = sdc.pPath
		}
	}
	if len(sdc.path) > 0 {
		if len(fp) > 0 {
			fp += "/" + sdc.path
		} else {
			fp = sdc.path
		}
	}
	return fp
}

func (sdc *sideCtx) fullPath() string {
	fp := sdc.root
	rp := sdc.relPath()
	if len(rp) > 0 {
		if len(fp) > 0 {
			fp += "/" + rp
		} else {
			fp = rp
		}
	}
	return fp
}

func (syc *syncCtx) pErr() error {
	if syc.err != nil {
		return fmt.Errorf("parent error %v", syc.err)
	}
	return nil
}

func (syc *syncCtx) cmpMeta(isRight bool) (updated, mUpdated bool) {
	lMeta, _ := syc.left.meta.(cabridss.Meta)
	rMeta, _ := syc.right.meta.(cabridss.Meta)
	if lMeta.Size != rMeta.Size || (lMeta.Ch != "" && lMeta.Ch != rMeta.Ch) {
		updated = true
		return
	}
	if lMeta.Mtime != rMeta.Mtime {
		mUpdated = true
		return
	}
	if syc.options.NoACL {
		return
	}
	if !isRight {
		mACL := syc.mapACL(lMeta.ACL, false)
		mUpdated = !cabridss.CmpAcl(rMeta.ACL, mACL)
	} else {
		mACL := syc.mapACL(rMeta.ACL, true)
		mUpdated = !cabridss.CmpAcl(lMeta.ACL, mACL)
	}
	return
}

func (syc *syncCtx) eval(rent *SyncReportEntry) {
	if syc.left.exist {
		if syc.right.exist {
			if syc.options.BiDir {
				rent.isRTL = syc.right.meta.GetMtime() > syc.left.meta.GetMtime()
			}
			rent.Updated, rent.MUpdated = syc.cmpMeta(rent.isRTL)
		} else {
			rent.Created = true
		}
	} else if syc.right.exist {
		if syc.options.BiDir {
			rent.isRTL = true
			rent.Created = true
		} else {
			if syc.options.KeepContent {
				rent.Kept = true
			} else {
				rent.Removed = true
			}
		}
	}
}

func (syc *syncCtx) makeChild(path string) syncCtx {
	npath := path
	isNs := path[len(path)-1] == '/'
	if isNs {
		npath = path[:len(path)-1]
	}

	return syncCtx{
		options: syc.options,
		err:     syc.pErr(),
		left: sideCtx{
			options: syc.options, dss: syc.left.dss,
			root: syc.left.root, pPath: syc.left.relPath(), isNs: isNs, path: npath,
			exist: syc.left.exist && (internal.NpType(path)).ExistIn(syc.left.exCh)},
		right: sideCtx{
			options: syc.options, dss: syc.right.dss, isRight: true,
			root: syc.right.root, pPath: syc.right.relPath(), isNs: isNs, path: npath,
			exist: syc.right.exist && (internal.NpType(path)).ExistIn(syc.right.exCh)},
	}
}

func (syc *syncCtx) diagnose(label string, sdDsp bool) {
	if syc.options.BeVerbose == nil {
		return
	}
	syc.options.BeVerbose(2, fmt.Sprintf("%-10s %s %s", label, syc.left.fullPath(), syc.right.fullPath()))
	if !sdDsp {
		return
	}
	syc.left.diagnose("  ")
	syc.right.diagnose("  ")
}

func (sdc *sideCtx) diagnose(label string) {
	if sdc.options.BeVerbose == nil {
		return
	}
	f := "%s %c %s %d %d %s"
	level := 3
	if label != "" && label[0] != ' ' {
		level = 2
		f = "%-10s %c %s %d %d %s"
	}
	meta := sdc.meta
	if meta == nil {
		meta = cabridss.Meta{}
	}
	sdc.options.BeVerbose(level, fmt.Sprintf(f, label, sdc.arrow(), sdc.fullPath(), meta.GetSize(), meta.GetMtime(), meta.GetChUnsafe()))
}
