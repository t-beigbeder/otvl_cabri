package cabrisync

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"io"
)

func (sdc *sideCtx) lsnsMeta() (err error) {
	if sdc.exist {
		sdc.diagnose(">lsnsMeta")
		if sdc.meta, err = sdc.dss.GetMeta(cabridss.AppendSlashIf(sdc.fullPath()), !sdc.options.NoCh); err != nil {
			sdc.exist = false
			sdc.diagnose(fmt.Sprintf("<lsnsMeta %v", err))
			return fmt.Errorf("in lsnsMeta: %c%s %w", sdc.arrow(), sdc.fullPath(), err)
		}
		sdc.exCh = sdc.meta.GetChildren()
		sdc.actualMtime = sdc.meta.GetMtime()
		sdc.diagnose("<lsnsMeta")
	}
	return nil
}

func (sdc *sideCtx) getMeta() (err error) {
	if sdc.exist {
		sdc.diagnose(">getMeta")
		if sdc.meta, err = sdc.dss.GetMeta(sdc.fullPath(), !sdc.options.NoCh); err != nil {
			sdc.exist = false
			sdc.diagnose(fmt.Sprintf("<getMeta %v", err))
			return fmt.Errorf("in getMeta: %c%s %w", sdc.arrow(), sdc.fullPath(), err)
		}
		sdc.actualMtime = sdc.meta.GetMtime()
		sdc.diagnose("<getMeta")
	}
	return nil
}

func (syc *syncCtx) mapACL(oACL []cabridss.ACLEntry, isRight bool) []cabridss.ACLEntry {
	if syc.options.NoACL {
		return nil
	}
	var tACL []cabridss.ACLEntry
	for _, oae := range oACL {
		tae := oae
		for lu, ru := range syc.options.MapACL {
			if (isRight && ru == oae.User) || (!isRight && lu == oae.User) {
				if isRight {
					tae.User = lu
				} else {
					tae.User = ru
				}
			}
		}
		tACL = append(tACL, tae)
	}
	return tACL
}

func (syc *syncCtx) evalMergeNsMeta(rent SyncReportEntry) (mtime int64, lAcl, rAcl []cabridss.ACLEntry) {
	if (rent.isRTL && rent.Created) || syc.left.meta == nil {
		if syc.right.meta == nil {
			syc.err = fmt.Errorf("in evalMergeNsMeta: FIX")
		}
		mtime = syc.right.meta.GetMtime()
		lAcl = syc.mapACL(syc.right.meta.GetAcl(), true)
	} else {
		mtime = syc.left.meta.GetMtime()
	}
	if syc.left.meta != nil {
		rAcl = syc.mapACL(syc.left.meta.GetAcl(), false)
	}
	return
}

func (syc *syncCtx) mergeNsBefore(rent SyncReportEntry) {
	mtime, lAcl, rAcl := syc.evalMergeNsMeta(rent)
	if syc.err == nil &&
		((syc.left.exist && len(syc.leftMg) != len(syc.left.exCh)) || (!syc.left.exist && rent.isRTL && rent.Created)) {
		syc.err = syc.left.crUpNs(mtime, syc.leftMg, lAcl)
	}
	if syc.err == nil &&
		((syc.right.exist && len(syc.rightMg) != len(syc.right.exCh)) || (!syc.right.exist && !rent.isRTL && rent.Created)) {
		syc.err = syc.right.crUpNs(mtime, syc.rightMg, rAcl)
	}
}

func (syc *syncCtx) mergeNsAfter(rent SyncReportEntry) {
	mtime, lAcl, rAcl := syc.evalMergeNsMeta(rent)
	if syc.err == nil &&
		((syc.left.exist && (rent.Updated || rent.MUpdated)) ||
			(!syc.left.exist && rent.isRTL && rent.Created)) {
		syc.err = syc.left.crUpNs(mtime, syc.leftMg, lAcl)
	}
	if syc.err == nil &&
		((syc.right.exist && (rent.Updated || rent.MUpdated)) ||
			(!syc.right.exist && !rent.isRTL && rent.Created)) {
		syc.err = syc.right.crUpNs(mtime, syc.leftRight, rAcl)
	}
}

func (sdc *sideCtx) crUpNs(mtime int64, children []string, acl []cabridss.ACLEntry) (err error) {
	if sdc.exist {
		sdc.diagnose(">crUpNsU")
		if err = sdc.dss.Updatens(sdc.fullPath(), mtime, children, acl); err != nil {
			sdc.diagnose(fmt.Sprintf("<crUpNs %v", err))
			return fmt.Errorf("in crUpNs: %c%s %w", sdc.arrow(), sdc.fullPath(), err)
		}
	} else {
		sdc.diagnose(">crUpNsM")
		if err = sdc.dss.Mkns(sdc.fullPath(), mtime, children, acl); err != nil {
			sdc.diagnose(fmt.Sprintf("<crUpNs %v", err))
			return fmt.Errorf("in crUpNs: %c%s %w", sdc.arrow(), sdc.fullPath(), err)
		}
		sdc.created = true
	}
	sdc.actualMtime = mtime
	sdc.diagnose("<crUpNs")
	return nil
}

func (syc *syncCtx) crUpContent(isRTL bool) error {
	syc.diagnose(">crUpContent", false)
	ori := syc.left
	tgt := syc.right
	if isRTL {
		ori = syc.right
		tgt = syc.left
	}
	var oErrPrefix = fmt.Sprintf("in crUpContent: %c%s", ori.arrow(), ori.fullPath())
	var tErrPrefix = fmt.Sprintf("in crUpContent: %c%s", tgt.arrow(), tgt.fullPath())

	in, err := ori.dss.GetContentReader(ori.fullPath())
	if err != nil {
		syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
		return fmt.Errorf("%s %w", oErrPrefix, err)
	}
	defer in.Close()
	var closeErr error
	out, err := tgt.dss.GetContentWriter(
		tgt.fullPath(), ori.meta.GetMtime(), syc.mapACL(ori.meta.GetAcl(), isRTL),
		func(err error, size int64, ch string) {
			if err != nil || size != ori.meta.GetSize() || (ori.meta.GetChUnsafe() != "" && ch != ori.meta.GetChUnsafe()) {
				closeErr = fmt.Errorf("%s error %w size %d ch %s", tErrPrefix, err, size, ch)
			}
		})
	if err != nil {
		err = fmt.Errorf("%s %w", tErrPrefix, err)
		syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		err = fmt.Errorf("%s %w", tErrPrefix, err)
		syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
		return err
	}
	if err = out.Close(); err != nil {
		err = fmt.Errorf("%s %w", tErrPrefix, err)
		syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
		return err
	}
	if closeErr == nil {
		if !tgt.exist {
			tgt.created = true
		}
		tgt.actualMtime = ori.meta.GetMtime()
		syc.diagnose("<crUpContent", false)
		return nil
	} else {
		syc.diagnose(fmt.Sprintf("<crUpContent %v", closeErr), false)
		return closeErr
	}
}
