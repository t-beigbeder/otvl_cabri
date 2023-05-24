package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/tidwall/buntdb"
	"os"
	"strings"
	"sync"
	"time"
)

type TimedMeta struct {
	Time  int64  `json:"time,string"`
	Bytes string `json:"bytes"`
}

type UpdatedData struct {
	Changed map[string][]TimedMeta `json:"changed"`
	Deleted map[string]bool        `json:"deleted"`
}

type Index interface {
	queryMetaTimes(npath string) ([]int64, error, bool)
	storeMetaTimes(npath string, times []int64) error
	loadMeta(npath string, time int64) ([]byte, error, bool)
	storeMeta(npath string, time int64, bs []byte) error
	removeMeta(npath string, time int64) error
	IsPersistent() bool
	isClientKnown(clId string) (bool, error)
	recordClient(clId string) (UpdatedData, error)
	updateClient(clId string, isFull bool) (UpdatedData, error)
	updateData(data UpdatedData, isFull bool) error
	Close() error
	Repair(readOnly bool) ([]string, error)
	Dump() string
}

type mIndex struct {
	metaTimes map[string][]int64
	metas     map[string][]byte
	clients   map[string]bool
	lock      sync.Mutex
}

func (mix *mIndex) queryMetaTimes(npath string) ([]int64, error, bool) {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	metaTimes, ok := mix.metaTimes[internal.NameToHashStr32(npath)]
	return metaTimes, nil, ok
}

func (mix *mIndex) storeMetaTimes(npath string, times []int64) error {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	mix.metaTimes[internal.NameToHashStr32(npath)] = times
	return nil
}

func (mix *mIndex) loadMeta(npath string, time int64) ([]byte, error, bool) {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	key := fmt.Sprintf("meta-%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(time))
	meta, ok := mix.metas[key]
	return meta, nil, ok
}

func (mix *mIndex) storeMeta(npath string, time int64, bs []byte) error {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	h := internal.NameToHashStr32(npath)
	found := false
	for _, ot := range mix.metaTimes[h] {
		if ot == time {
			found = true
			break
		}
	}
	if !found {
		mix.metaTimes[h] = append(mix.metaTimes[h], time)
	}
	key := fmt.Sprintf("meta-%s.%s", h, internal.Int64ToStr16(time))
	mix.metas[key] = bs
	return nil
}

func (mix *mIndex) removeMeta(npath string, time int64) error {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	h := internal.NameToHashStr32(npath)
	key := fmt.Sprintf("meta-%s.%s", h, internal.Int64ToStr16(time))
	if _, ok := mix.metas[key]; !ok {
		return fmt.Errorf("in removeMeta: %s not found", key)
	}
	found := false
	var mts []int64
	for _, ot := range mix.metaTimes[h] {
		if ot == time {
			found = true
		} else {
			mts = append(mts, ot)
		}
	}
	if !found {
		return fmt.Errorf("in removeMeta: %s, %d not found", key, time)
	}
	delete(mix.metas, key)
	if len(mts) != 0 {
		mix.metaTimes[h] = mts
	} else {
		delete(mix.metaTimes, h)
	}
	return nil
}

func (mix *mIndex) Close() error { return nil }

func (mix *mIndex) IsPersistent() bool { return false }

func (mix *mIndex) isClientKnown(clId string) (bool, error) {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	_, ok := mix.clients[clId]
	return ok, nil
}

func (mix *mIndex) recordClient(clId string) (UpdatedData, error) {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	mix.clients[clId] = true
	return UpdatedData{}, nil
}

func (mix *mIndex) updateClient(clId string, isFull bool) (UpdatedData, error) {
	mix.lock.Lock()
	defer mix.lock.Unlock()
	_, ok := mix.clients[clId]
	if !ok {
		return UpdatedData{}, fmt.Errorf("in updateClient: %s unknown", clId)
	}
	return UpdatedData{}, nil
}

func (mix *mIndex) updateData(data UpdatedData, isFull bool) error { panic("not implemented") }

func (mix *mIndex) Repair(readOnly bool) ([]string, error) { return nil, nil }

func (mix *mIndex) Dump() string { return "" }

func NewMIndex() Index {
	return &mIndex{metaTimes: map[string][]int64{}, metas: map[string][]byte{}, clients: map[string]bool{}}
}

type PixClient struct {
	InternalId int `json:"internalId"` // client internal id (key prefix to client entries in DB)
	TxId       int `json:"txCounter"`  // current client transaction id (key prefix to client new transactions in DB)
}

type PixClients struct {
	IdCounter int                   `json:"counter"` // next internal id to be used
	Clients   map[string]*PixClient `json:"clients"` // mapping external client id to internal id
}

type pIndex struct {
	path    string
	db      *buntdb.DB
	clients PixClients
	closed  bool
}

func (pix *pIndex) queryMetaTimes(npath string) (metaTimes []int64, err error, ok bool) {
	err = pix.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(fmt.Sprintf("mts/%s", internal.NameToHashStr32(npath)))
		if err == buntdb.ErrNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		ok = true
		if val == "" {
			return nil
		}
		for _, mt := range strings.Split(val, " ") {
			it, err := internal.Str16ToInt64(mt)
			if err != nil {
				return err
			}
			metaTimes = append(metaTimes, it)
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("in queryMetaTimes: %v", err)
	}
	return
}

func (pix *pIndex) doStoreMetaTimes(tx *buntdb.Tx, nph string, smts string) error {
	_, _, err := tx.Set(fmt.Sprintf("mts/%s", nph), smts, nil)
	if err != nil {
		return err
	}
	for _, pc := range pix.clients.Clients {
		_, _, err := tx.Set(fmt.Sprintf("%d/%012d/mts/%s", pc.InternalId, pc.TxId, nph), smts, nil)
		if err != nil {
			return err
		}
		if smts != "" {
			continue
		}
		_, _, err = tx.Set(fmt.Sprintf("%d/%012d/xmts/%s", pc.InternalId, pc.TxId, nph), smts, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func ts2sts(times []int64) string {
	var elems []string
	for _, mt := range times {
		elems = append(elems, internal.Int64ToStr16(mt))
	}
	return strings.Join(elems, " ")
}

func (pix *pIndex) storeMetaTimes(npath string, times []int64) error {
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		return pix.doStoreMetaTimes(tx, internal.NameToHashStr32(npath), ts2sts(times))
	})
	if err != nil {
		err = fmt.Errorf("in storeMetaTimes: %v", err)
	}
	return err
}

func (pix *pIndex) loadMeta(npath string, time int64) (meta []byte, err error, ok bool) {
	pix.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(
			fmt.Sprintf("m/%s.%s", internal.NameToHashStr32(npath), internal.Int64ToStr16(time)))
		if err == buntdb.ErrNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		ok = true
		meta = []byte(val)
		return nil
	})
	if err != nil {
		err = fmt.Errorf("in loadMeta: %v", err)
	}
	return
}

func (pix *pIndex) doStoreMeta(tx *buntdb.Tx, nph string, time int64, bs []byte) error {
	st := internal.Int64ToStr16(time)
	val, err := tx.Get(fmt.Sprintf("mts/%s", nph))
	if err != nil && err != buntdb.ErrNotFound {
		return err
	}
	found := false
	for _, mt := range strings.Split(val, " ") {
		if mt == st {
			found = true
			break
		}
	}
	if !found {
		if val != "" {
			val = strings.Join(append(strings.Split(val, " "), st), " ")
		} else {
			val = st
		}
		if err := pix.doStoreMetaTimes(tx, nph, val); err != nil {
			return err
		}
	}
	if _, _, err = tx.Set(fmt.Sprintf("m/%s.%s", nph, internal.Int64ToStr16(time)), string(bs), nil); err != nil {
		return err
	}
	return nil
}

func (pix *pIndex) storeMeta(npath string, time int64, bs []byte) error {
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		nph := internal.NameToHashStr32(npath)
		return pix.doStoreMeta(tx, nph, time, bs)
	})
	if err != nil {
		err = fmt.Errorf("in storeMeta: %v", err)
	}
	return err
}

func (pix *pIndex) doRemoveMeta(tx *buntdb.Tx, nph string, time int64) error {
	st := internal.Int64ToStr16(time)
	val, err := tx.Get(fmt.Sprintf("mts/%s", nph))
	if err != nil {
		return err
	}
	found := false
	var smts []string
	for _, mt := range strings.Split(val, " ") {
		if mt == st {
			found = true
		} else {
			smts = append(smts, mt)
		}
	}
	if !found {
		return fmt.Errorf("in doRemoveMeta: %s %d %s not found", nph, time, st)
	}
	mkey := fmt.Sprintf("m/%s.%s", nph, internal.Int64ToStr16(time))
	if _, err := tx.Get(mkey); err != nil {
		return fmt.Errorf("in doRemoveMeta: %s not found", mkey)
	}
	if err := pix.doStoreMetaTimes(tx, nph, strings.Join(smts, " ")); err != nil {
		return err
	}
	if _, err = tx.Delete(mkey); err != nil {
		return fmt.Errorf("in doRemoveMeta: %v", err)
	}
	return nil
}

func (pix *pIndex) removeMeta(npath string, time int64) error {
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		nph := internal.NameToHashStr32(npath)
		return pix.doRemoveMeta(tx, nph, time)
	})
	if err != nil {
		err = fmt.Errorf("in removeMeta: %s %v", npath, err)
	}
	return err
}

func (pix *pIndex) Close() error {
	if !pix.closed {
		pix.closed = true
		err := pix.db.Update(func(tx *buntdb.Tx) error {
			_, err := tx.Delete("g/lock")
			return err
		})
		if err != nil {
			pix.db.Close()
			return fmt.Errorf("in pIndex.Close: %v", err)
		}
	}
	return pix.db.Close()
}

func (pix *pIndex) IsPersistent() bool { return true }

func (pix *pIndex) doPurgeClient(tx *buntdb.Tx, clId string, isFull bool) error {
	for txId := 0; txId <= pix.clients.Clients[clId].TxId; txId++ {
		if !isFull && txId == pix.clients.Clients[clId].TxId {
			continue
		}
		var purgedKeys []string
		prefix := fmt.Sprintf("%d/%012d/", pix.clients.Clients[clId].InternalId, txId)
		tx.Ascend("", func(key, value string) bool {
			if !strings.HasPrefix(key, prefix) {
				return true
			}
			purgedKeys = append(purgedKeys, key)
			return true
		})
		for _, key := range purgedKeys {
			if _, err := tx.Delete(key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (pix *pIndex) doUpdateClient(tx *buntdb.Tx, clId string, isFull bool) (UpdatedData, error) {
	udd := UpdatedData{Changed: map[string][]TimedMeta{}, Deleted: map[string]bool{}}
	updateData := func(nph string, mts string) error {
		tms, _ := udd.Changed[nph]
		if mts == "" {
			return nil
		}
		for _, mt := range strings.Split(mts, " ") {
			mk := fmt.Sprintf("m/%s.%s", nph, mt)
			val, err := tx.Get(mk)
			if err != nil {
				return fmt.Errorf("key %s: %v", mk, err)
			}
			it, _ := internal.Str16ToInt64(mt)
			tms = append(tms, TimedMeta{Time: it, Bytes: val})
		}
		udd.Changed[nph] = tms
		return nil
	}

	var err, intErr error

	if isFull {
		if err = pix.doPurgeClient(tx, clId, true); err != nil {
			return UpdatedData{}, err
		}
		err = tx.Ascend("", func(key, value string) bool {
			if !strings.HasPrefix(key, "mts/") {
				return true
			}
			if intErr = updateData(key[len("mts/"):], value); intErr != nil {
				return false
			}
			return true
		})
	} else {
		prefix := fmt.Sprintf("%d/%012d/mts/", pix.clients.Clients[clId].InternalId, pix.clients.Clients[clId].TxId)
		xPrefix := fmt.Sprintf("%d/%012d/xmts/", pix.clients.Clients[clId].InternalId, pix.clients.Clients[clId].TxId)
		err = tx.Ascend("", func(key, value string) bool {
			if strings.HasPrefix(key, prefix) {
				if intErr = updateData(key[len(prefix):], value); intErr != nil {
					return false
				}
			} else if strings.HasPrefix(key, xPrefix) {
				udd.Deleted[key[len(prefix):]] = true
			}
			return true
		})
		if err == nil {
			if err = pix.doPurgeClient(tx, clId, false); err != nil {
				return UpdatedData{}, err
			}
		}
	}
	if intErr != nil {
		err = intErr
	}
	if err != nil {
		return UpdatedData{}, fmt.Errorf("in doUpdateClient: %v", err)
	}

	if isFull {
		pix.clients.Clients[clId].TxId = 0
	} else {
		pix.clients.Clients[clId].TxId++
	}
	bsCls, err := json.Marshal(pix.clients)
	if err != nil {
		return UpdatedData{}, fmt.Errorf("in doUpdateClient: %v", err)
	}
	if _, _, err = tx.Set("g/clients", string(bsCls), nil); err != nil {
		return UpdatedData{}, fmt.Errorf("in doUpdateClient: %v", err)
	}

	return udd, nil
}

func (pix *pIndex) isClientKnown(clId string) (bool, error) {
	_, ok := pix.clients.Clients[clId]
	return ok, nil
}

func (pix *pIndex) recordClient(clId string) (UpdatedData, error) {
	udd := UpdatedData{}
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		var err error
		if _, ok := pix.clients.Clients[clId]; ok {
			return fmt.Errorf("client %s already recorded", clId)
		}
		pix.clients.Clients[clId] = &PixClient{InternalId: pix.clients.IdCounter}
		pix.clients.IdCounter += 1
		if udd, err = pix.doUpdateClient(tx, clId, true); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("in recordClient: %v", err)
	}
	return udd, err
}

func (pix *pIndex) updateClient(clId string, isFull bool) (UpdatedData, error) {
	var udd UpdatedData
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		var err error
		if _, ok := pix.clients.Clients[clId]; !ok {
			return fmt.Errorf("client %s has not been recorded", clId)
		}
		udd, err = pix.doUpdateClient(tx, clId, isFull)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("in updateClient: %v", err)
	}
	return udd, err
}

func (pix *pIndex) updateData(udd UpdatedData, isFull bool) error {
	err := pix.db.Update(func(tx *buntdb.Tx) error {
		if isFull {
			var purgedKeys []string
			tx.Ascend("", func(key, value string) bool {
				if !strings.HasPrefix(key, "g/") {
					purgedKeys = append(purgedKeys, key)
				}
				return true
			})
			for _, key := range purgedKeys {
				if _, err := tx.Delete(key); err != nil {
					return err
				}
			}
		}

		for nph, tms := range udd.Changed {
			updTms := false
			var eTimes []string
			if !isFull {
				sETimes, err := tx.Get(fmt.Sprintf("mts/%s", nph))
				if err != nil && err != buntdb.ErrNotFound {
					return err
				}
				if sETimes != "" {
					eTimes = strings.Split(sETimes, " ")
				}
			}
			for _, tm := range tms {
				sTime := internal.Int64ToStr16(tm.Time)
				found := false
				for _, eTime := range eTimes {
					if eTime == sTime {
						found = true
						break
					}
				}
				if !found {
					updTms = true
					eTimes = append(eTimes, sTime)
				}
				if _, _, err := tx.Set(fmt.Sprintf("m/%s.%s", nph, sTime), string(tm.Bytes), nil); err != nil {
					return err
				}

			}
			if updTms {
				if _, _, err := tx.Set(fmt.Sprintf("mts/%s", nph), strings.Join(eTimes, " "), nil); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("in updateData: %v", err)
	}
	return nil
}

func (pix *pIndex) Dump() string {
	var lines []string
	pix.db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
			return true
		})
	})
	return strings.Join(lines, "\n")
}

func loadInMemory(db *buntdb.DB) (map[string]map[int64]bool, map[string]map[int64][]byte, map[string]bool, error) {
	metaTimes := map[string]map[int64]bool{}
	metas := map[string]map[int64][]byte{}
	removed := map[string]bool{}
	scanMeta := func(key, value string) {
		frags := strings.Split(key[2:], ".")
		if len(frags) != 2 {
			removed[key] = true
			return
		}
		i, err := internal.Str16ToInt64(frags[1])
		if err != nil {
			removed[key] = true
			return
		}
		if _, ok := metas[frags[0]]; !ok {
			metas[frags[0]] = map[int64][]byte{}
		}
		metas[frags[0]][i] = []byte(value)
		return
	}
	scanMts := func(key, value string, isClient bool) {
		if isClient {
			frags := strings.Split(key, "/")
			if len(frags) != 4 {
				removed[key] = true
				return
			}
		}
		smts := strings.Split(value, " ")
		if len(smts) == 0 {
			removed[key] = true
			return
		}
		for _, smt := range smts {
			mt, err := internal.Str16ToInt64(smt)
			if err != nil {
				removed[key] = true
				continue
			}
			if _, ok := metaTimes[key]; !ok {
				metaTimes[key] = map[int64]bool{}
			}
			metaTimes[key][mt] = true
		}
	}
	if err := db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			switch {
			case strings.HasPrefix(key, "m/"):
				scanMeta(key, value)
			case strings.HasPrefix(key, "mts/"):
				scanMts(key, value, false)
			case strings.Contains(key, "/mts/"):
				scanMts(key, value, true)
			}
			return true
		})
	}); err != nil {
		return nil, nil, nil, err
	}
	return metaTimes, metas, removed, nil
}

func (pix *pIndex) loadInMemory() (map[string]map[int64]bool, map[string]map[int64][]byte, map[string]bool, error) {
	return loadInMemory(pix.db)
}

func doRepair(db *buntdb.DB, readOnly bool) ([]string, map[string][]AuditIndexInfo, error) {
	metaTimes, metas, removed, err := loadInMemory(db)
	if err != nil {
		return nil, nil, fmt.Errorf("in doRepair: %v", err)
	}

	getNph := func(key string) string {
		switch {
		case strings.HasPrefix(key, "mts/"):
			return key[4:]
		case strings.Contains(key, "/mts/"):
			return strings.Split(key, "/")[3]
		}
		return ""
	}

	updated := map[string]bool{}
	mai := map[string][]AuditIndexInfo{}
	appMai := func(key string, aii AuditIndexInfo) {
		aiis, _ := mai[key]
		mai[key] = append(aiis, aii)
	}
	ds := []string{}
	for key, mts := range metaTimes {
		nph := getNph(key)
		if _, ok := metas[nph]; !ok {
			ds = append(ds, fmt.Sprintf("x %s", key))
			removed[key] = true
			appMai(nph, AuditIndexInfo{Time: MIN_TIME, Error: "IndexInternal"})
			continue
		}
		for mt, _ := range mts {
			if _, ok := metas[nph][mt]; !ok {
				ds = append(ds, fmt.Sprintf("* %s", key))
				updated[key] = true
				appMai(nph, AuditIndexInfo{Time: mt, Error: "IndexInternal"})
			}
		}
	}
	if readOnly {
		return ds, mai, nil
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		for key, _ := range removed {
			if _, err := tx.Delete(key); err != nil {
				return err
			}
		}
		for key, _ := range updated {
			nph := getNph(key)
			mts := metaTimes[key]
			newMts := []int64{}
			for mt, _ := range mts {
				if _, ok := metas[nph][mt]; ok {
					newMts = append(newMts, mt)
				}
			}
			if len(newMts) > 0 {
				if _, _, err := tx.Set(key, ts2sts(newMts), nil); err != nil {
					return err
				}
			} else {
				if _, err := tx.Delete(key); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return ds, mai, err
}

func (pix *pIndex) pRepair() (map[string][]AuditIndexInfo, error) {
	_, mai, err := doRepair(pix.db, true)
	return mai, err
}

func (pix *pIndex) Repair(readOnly bool) ([]string, error) {
	ds, _, err := doRepair(pix.db, readOnly)
	return ds, err
}

func bdbReconfigure(db *buntdb.DB) error {
	var config buntdb.Config
	if err := db.ReadConfig(&config); err != nil {
		return err
	}
	config.SyncPolicy = buntdb.Never
	if err := db.SetConfig(config); err != nil {
		return err
	}
	return nil
}

func NewPIndex(path string, unlock, autoRepair bool) (Index, error) {
	db, err := buntdb.Open(path)
	var (
		clients  PixClients
		unlocked bool
	)
	if err != nil {
		return nil, fmt.Errorf("in NewPIndex: %v", err)
	}
	if err = bdbReconfigure(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("in NewPIndex: %v", err)
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		previous, replaced, err := tx.Set("g/lock", time.Now().Format("2006-01-02 15:04:05.000"), nil)
		if err != nil {
			return err
		}
		if replaced {
			if !unlock {
				return fmt.Errorf("index %s locked since %s", path, previous)
			}
			unlocked = true
		}
		sCls, err := tx.Get("g/clients")
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		if err == nil {
			if err = json.Unmarshal([]byte(sCls), &clients); err != nil {
				return err
			}
		} else {
			clients.Clients = make(map[string]*PixClient)
			bsCls, err := json.Marshal(clients)
			if err != nil {
				return err
			}
			if _, _, err = tx.Set("g/clients", string(bsCls), nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("in NewPIndex: %v", err)
	}
	if unlocked && autoRepair {
		if _, _, err = doRepair(db, false); err != nil {
			return nil, fmt.Errorf("in NewPIndex: %v", err)
		}
	}
	return &pIndex{path: path, db: db, clients: clients}, nil
}

func reindexPIndex(path string, metaTimes map[string]map[int64]bool, metas map[string]map[int64][]byte) error {
	index, err := NewPIndex(path, true, false)
	if err != nil {
		return err
	}
	index.Close()
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("in reindexPIndex: %w", err)
	}
	db, err := buntdb.Open(path)
	if err != nil {
		return fmt.Errorf("in reindexPIndex: %w", err)
	}
	db.Close()
	index, err = NewPIndex(path, false, false)
	if err != nil {
		return err
	}
	index.Close()
	db, err = buntdb.Open(path)
	if err != nil {
		return fmt.Errorf("in reindexPIndex: %w", err)
	}
	if err = bdbReconfigure(db); err != nil {
		db.Close()
		return fmt.Errorf("in reindexPIndex: %v", err)
	}
	defer db.Close()
	err = db.Update(func(tx *buntdb.Tx) error {
		for nph, mts := range metaTimes {
			var newMts []int64
			for mt, _ := range mts {
				newMts = append(newMts, mt)
				sTime := internal.Int64ToStr16(mt)
				mb := metas[nph][mt]
				if _, _, err := tx.Set(fmt.Sprintf("m/%s.%s", nph, sTime), string(mb), nil); err != nil {
					return err
				}
			}
			if _, _, err := tx.Set(fmt.Sprintf("mts/%s", nph), ts2sts(newMts), nil); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

type nIndex struct{}

func (n *nIndex) queryMetaTimes(npath string) ([]int64, error, bool) { return nil, nil, false }

func (n *nIndex) storeMetaTimes(npath string, times []int64) error { return nil }

func (n *nIndex) loadMeta(npath string, time int64) ([]byte, error, bool) { return nil, nil, false }

func (n *nIndex) storeMeta(npath string, time int64, bs []byte) error { return nil }

func (n *nIndex) removeMeta(npath string, time int64) error { return nil }

func (n *nIndex) Close() error { return nil }

func (n *nIndex) IsPersistent() bool { return false }

func (n *nIndex) isClientKnown(clId string) (bool, error) { panic("not implemented") }

func (n *nIndex) recordClient(clId string) (UpdatedData, error) { panic("not implemented") }

func (n *nIndex) updateClient(clId string, isFull bool) (UpdatedData, error) {
	panic("not implemented")
}

func (n *nIndex) updateData(data UpdatedData, isFull bool) error { panic("not implemented") }

func (n *nIndex) Dump() string { return "" }

func (n *nIndex) Repair(readOnly bool) ([]string, error) { return nil, nil }

func NewNIndex() Index { return &nIndex{} }
