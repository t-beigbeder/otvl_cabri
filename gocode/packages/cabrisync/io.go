package cabrisync

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"io"
	"strings"
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

func mapACE(oace cabridss.ACLEntry, tu string, rightsMask cabridss.Rights) cabridss.ACLEntry {
	return cabridss.ACLEntry{
		User: tu,
		Rights: cabridss.Rights{
			Read:    oace.Rights.Read && rightsMask.Read,
			Write:   oace.Rights.Write && rightsMask.Write,
			Execute: oace.Rights.Execute && rightsMask.Execute,
		},
	}
}

func appendAceIf(tacl []cabridss.ACLEntry, ace cabridss.ACLEntry, hasMeta bool) []cabridss.ACLEntry {
	if !hasMeta && ace.User != "" && !strings.HasPrefix(ace.User, "x-uid:") &&
		!strings.HasPrefix(ace.User, "x-gid:") && ace.User != "x-other" {
		return tacl
	}
	for _, tace := range tacl {
		if tace.User == ace.User {
			return tacl
		}
	}
	return append(tacl, ace)
}

func (syc *syncCtx) mapACL(oACL []cabridss.ACLEntry, isRight bool) []cabridss.ACLEntry {
	if syc.options.NoACL {
		return nil
	}
	var tacl []cabridss.ACLEntry
	for _, oace := range oACL {
		macl := map[string][]cabridss.ACLEntry{}
		var hasMeta bool
		if !isRight {
			_, hasMeta = syc.right.dss.(cabridss.HDss)
			macl = syc.options.LeftMapACL
		} else {
			_, hasMeta = syc.left.dss.(cabridss.HDss)
			macl = syc.options.RightMapACL
		}
		done := false
		for cou, cmacl := range macl {
			if cou == oace.User {
				done = true
				for _, mace := range cmacl {
					tacl = appendAceIf(tacl, mapACE(oace, mace.User, mace.Rights), hasMeta)
				}
			}
		}
		if !done {
			tacl = appendAceIf(tacl, mapACE(oace, oace.User, cabridss.Rights{Execute: true, Read: true, Write: true}), hasMeta)
		}
	}
	return tacl
}

func (syc *syncCtx) evalMergeNsMeta(rent SyncReportEntry) (mtime int64, lAcl, rAcl []cabridss.ACLEntry) {
	if (rent.isRTL && rent.Created) || syc.left.meta == nil {
		if syc.right.meta == nil {
			panic(fmt.Sprintf("evalMergeNsMeta %+v", rent))
		}
		mtime = syc.right.meta.GetMtime()
		lAcl = syc.mapACL(syc.right.meta.GetAcl(), true)
		rAcl = syc.right.meta.GetAcl()
	} else {
		if syc.left.meta == nil {
			panic(fmt.Sprintf("evalMergeNsMeta %+v", rent))
		}
		mtime = syc.left.meta.GetMtime()
		lAcl = syc.left.meta.GetAcl()
		rAcl = syc.mapACL(syc.left.meta.GetAcl(), false)
	}
	return
}

func (syc *syncCtx) mergeNsBefore(rent SyncReportEntry) {
	mtime, lAcl, rAcl := syc.evalMergeNsMeta(rent)
	if syc.err == nil &&
		((syc.left.exist && len(syc.leftMg) != len(syc.left.exCh)) || (!syc.left.exist && rent.isRTL && rent.Created)) {
		if syc.left.meta != nil {
			syc.err = syc.left.dss.SuEnableWrite(syc.left.meta.GetPath())
		}
		if syc.err == nil {
			syc.err = syc.left.crUpNs(mtime, syc.leftMg, lAcl)
		}
	}
	if syc.err == nil &&
		((syc.right.exist && len(syc.rightMg) != len(syc.right.exCh)) || (!syc.right.exist && !rent.isRTL && rent.Created)) {
		if syc.right.meta != nil {
			syc.err = syc.right.dss.SuEnableWrite(syc.right.meta.GetPath())
		}
		if syc.err == nil {
			syc.err = syc.right.crUpNs(mtime, syc.rightMg, rAcl)
		}
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
	if tgt.meta != nil {
		err := tgt.dss.SuEnableWrite(tgt.meta.GetPath())
		if err != nil {
			err = fmt.Errorf("%s %w", tErrPrefix, err)
			syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
			return err
		}
	}

	in, err := ori.dss.GetContentReader(ori.fullPath())
	if err != nil {
		syc.diagnose(fmt.Sprintf("<crUpContent %v", err), false)
		return fmt.Errorf("%s %w", oErrPrefix, err)
	}
	defer in.Close()
	var closeErr error
	var out io.WriteCloser
	doCopy := func() error {
		var err error
		out, err = tgt.dss.GetContentWriter(
			tgt.fullPath(), ori.meta.GetMtime(), syc.mapACL(ori.meta.GetAcl(), isRTL),
			func(err error, size int64, ch string) {
				if err != nil || size != ori.meta.GetSize() || (ori.meta.GetChUnsafe() != "" && ch != ori.meta.GetChUnsafe()) {
					closeErr = fmt.Errorf("%s error %w size %d ch %s", tErrPrefix, err, size, ch)
				}
			})
		if err != nil {
			err = fmt.Errorf("%s %w", tErrPrefix, err)
			return err
		}
		if _, err = io.Copy(out, in); err != nil {
			out.Close()
			err = fmt.Errorf("%s %w", tErrPrefix, err)
			return err
		}
		if err = out.Close(); err != nil {
			err = fmt.Errorf("%s %w", tErrPrefix, err)
			return err
		}
		return nil
	}
	if err = doCopy(); err != nil {
		if strings.Contains(err.Error(), "connect: cannot assign requested address") {
			in2, err2 := ori.dss.GetContentReader(ori.fullPath())
			if err2 != nil {
				in.Close()
				syc.diagnose(fmt.Sprintf("<crUpContent %v", err2), false)
				return fmt.Errorf("%s %w", oErrPrefix, err2)
			}
			in.Close()
			in = in2
			err = doCopy()
		}
	}
	if err != nil {
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
