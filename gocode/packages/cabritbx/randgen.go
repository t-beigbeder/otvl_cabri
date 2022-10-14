package cabritbx

import (
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type RandGenConfig struct {
	Seed                                                  int64
	MaxDepth, MaxEntries, NsRate, CrRate, RmRate, KmtRate int
	TimeOrigin                                            int64
}

func GetDefaultConfig() RandGenConfig {
	return RandGenConfig{
		Seed:       42,
		MaxDepth:   7,
		MaxEntries: 128,
		NsRate:     1,
		CrRate:     10,
		RmRate:     10,
		KmtRate:    1,
		TimeOrigin: time.Date(2002, time.January, 8, 18, 52, 0, 0, time.UTC).Unix(),
	}
}

type RandGen interface {
	Create(draws int) error
	Update(draws int) error
	CurTime() int64
	AdvTime(delta int64)
}

type randGen struct {
	RandGenConfig
	rand     *rand.Rand
	UpRate   int
	dss      cabridss.Dss
	entries  []string
	content  string
	isNs     bool
	curTime  int64
	isUp     bool
	isCr     bool
	isRm     bool
	doKmt    bool
	nps, cps []string
}

func NewRanGen(config RandGenConfig, dss cabridss.Dss) RandGen {
	return &randGen{
		RandGenConfig: config,
		curTime:       config.TimeOrigin,
		rand:          rand.New(rand.NewSource(config.Seed)),
		UpRate:        100 - config.CrRate - config.RmRate,
		dss:           dss}
}

func (rg *randGen) Create(draws int) error { return rg.crOrUp(draws, true) }

func (rg *randGen) Update(draws int) error { return rg.crOrUp(draws, false) }

func (rg *randGen) CurTime() int64 { return rg.curTime }

func (rg *randGen) AdvTime(delta int64) { rg.curTime += delta }

func (rg *randGen) drawEntries() (entries []string) {
	depth := rg.rand.Intn(rg.MaxDepth-1) + 1
	entries = make([]string, depth)
	for i := 0; i < depth; i++ {
		entries[i] = fmt.Sprintf("%d", rg.rand.Intn(rg.MaxEntries+1))
	}
	return
}

func (rg *randGen) lsnsRecur(npath string) (nps, cps []string, err error) {
	cs, err := rg.dss.Lsns(npath)
	if err != nil {
		return nil, nil, err
	}
	pp := npath
	if npath != "" {
		pp = pp + "/"
	}
	for _, c := range cs {
		if c[len(c)-1] == '/' {
			fc := fmt.Sprintf("%s%s", pp, c[:len(c)-1])
			nps = append(nps, fc)
			snps, scps, err := rg.lsnsRecur(fc)
			if err != nil {
				return nil, nil, err
			}
			nps = append(nps, snps...)
			cps = append(cps, scps...)
		} else {
			cps = append(cps, fmt.Sprintf("%s%s", pp, c))
		}
	}
	return
}

func (rg *randGen) crOrUp(draws int, isCreate bool) error {
	if !isCreate {
		var err error
		rg.nps, rg.cps, err = rg.lsnsRecur("")
		if err != nil {
			return err
		}
		sort.Strings(rg.nps)
		sort.Strings(rg.cps)
	}
	for draw := 0; draw < draws; draw++ {
		rg.isNs = rg.rand.Intn(1000) < rg.NsRate*10
		rg.entries = rg.drawEntries()
		if !rg.isNs {
			rg.content = strings.Join(rg.drawEntries(), "\n")
		}
		rg.curTime += int64(rg.rand.Intn(3597) + 3)
		rg.dss.SetCurrentTime(rg.curTime)

		if isCreate {
			if err := rg.doCreate(); err != nil {
				return err
			}
		} else {
			rg.isCr, rg.isRm, rg.isUp = false, false, false
			rg.doKmt = rg.rand.Intn(1000) < rg.KmtRate*10
			if rg.doKmt {
				rg.isUp = true
			} else {
				drw := rg.rand.Intn(1000)
				if drw < rg.CrRate*10 {
					rg.isCr = true
				} else if drw < (rg.CrRate+rg.RmRate)*10 {
					rg.isRm = true
				} else {
					rg.isUp = true
				}
			}

			if err := rg.doUpdate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rg *randGen) doCreate() error {
	npath := strings.Join(rg.entries, "/")
	if rg.isNs {
		if err := cabridss.MkallNs(rg.dss, npath, rg.curTime); err != nil {
			return err
		}
	} else {
		if err := cabridss.MkallContent(rg.dss, npath+".c", rg.curTime); err != nil {
			return err
		}
		cw, err := rg.dss.GetContentWriter(npath+".c", rg.curTime, nil, nil)
		if err != nil {
			return err
		}
		_, err = cw.Write([]byte(rg.content))
		if err != nil {
			cw.Close()
			return err
		}
		if err = cw.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (rg *randGen) removeSilently(path string) (err error) {
	parent := cabridss.Parent(path)
	meta, err := rg.dss.GetMeta(parent+"/", false)
	if err != nil {
		return err
	}
	if err = rg.dss.Remove(path); err != nil {
		return err
	}
	cs, err := rg.dss.Lsns(parent)
	if err != nil {
		return err
	}
	return rg.dss.Updatens(parent, meta.GetMtime(), cs, nil)
}

func (rg *randGen) doUpdate() error {
	if rg.isCr {
		return rg.doCreate()
	}
	if rg.isNs {
		npath := rg.nps[rg.rand.Intn(len(rg.nps))]
		meta, err := rg.dss.GetMeta(cabridss.AppendSlashIf(npath), false)
		if rg.isUp {
			cs := meta.GetChildren()
			if meta == nil || len(cs) == 0 {
				return cabridss.MkallNs(rg.dss, npath, rg.curTime)
			}
			if err != nil {
				return err
			}
			mtime := rg.curTime
			if rg.doKmt && meta != nil {
				mtime = meta.GetMtime()
			}
			if err = rg.dss.Updatens(npath, mtime, cs[:len(cs)-1], nil); err != nil {
				return err
			}
			return nil
		}
		if meta == nil || err != nil {
			return nil
		}
		if err = rg.removeSilently(npath + "/"); err != nil {
			return err
		}
		return nil
	}

	cpath := rg.cps[rg.rand.Intn(len(rg.cps))]
	if rg.isUp {
		meta, err := rg.dss.GetMeta(cpath, false)
		if meta == nil {
			if err = cabridss.MkallContent(rg.dss, cpath, rg.curTime); err != nil {
				return err
			}
		}
		mtime := rg.curTime
		if rg.doKmt && meta != nil {
			mtime = meta.GetMtime()
		}
		cw, err := rg.dss.GetContentWriter(cpath, mtime, nil, nil)
		if err != nil {
			return err
		}
		defer cw.Close()
		_, err = cw.Write([]byte(rg.content))
		if err != nil {
			return err
		}
		return nil
	}
	meta, err := rg.dss.GetMeta(cpath, false)
	if meta == nil {
		return nil
	}
	if err = rg.removeSilently(cpath); err != nil {
		return err
	}
	return nil
}
