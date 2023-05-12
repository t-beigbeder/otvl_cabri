package cabridss

import (
	"fmt"
	"net/http"
)

type webDssClientConfig struct {
	WebDssConfig
	libDss     HDss
	identities []IdentityConfig
}

type mInitialized struct {
	mError
	RepoId          string `json:"repoId"`
	PersistentIndex bool   `json:"persistentIndex"`
	Encrypted       bool   `json:"encrypted"`
	ClientIsKnown   bool   `json:"clientIsKnown"`
}

type mUpdatedData struct {
	mError
	UpdatedData
}

type mTimes struct {
	mError
	Times []int64 `json:"times,string"`
}

type mStoreMeta struct {
	Npath string     `json:"npath"`
	Time  int64      `json:"time,string"`
	Bs    []byte     `json:"bs,string"`
	ACL   []ACLEntry `json:"acl"`
}

type mRemoveMeta struct {
	Npath string `json:"npath"`
	Time  int64  `json:"time,string"`
}

type mPushContentIn struct {
	Size int64  `json:"size"`
	Ch   string `json:"ch"`
	Mbs  []byte `json:"bs,string"`
	Emid string `json:"emid"`
}

type mLoadMetaIn struct {
	Npath string `json:"npath"`
	Time  int64  `json:"time,string"`
}

type mLoadMetaOut struct {
	mError
	Bs []byte `json:"bs,string"`
}

type mSpGetContentReader struct {
	Ch string `json:"ch"`
}

type mExist struct {
	mError
	Exist bool `json:"exist"`
}

type mDump struct {
	mError
	Dump string `json:"dump"`
}

type mSPS struct {
	mError
	Sti  StorageInfo    `json:"sti"`
	Errs ErrorCollector `json:"errs"`
}

func aInitialize(clId string, dss HDss) *mInitialized {
	cik, err := dss.GetIndex().isClientKnown(clId)
	if err != nil {
		return &mInitialized{mError: mError{Error: err.Error()}}
	}
	return &mInitialized{
		RepoId:          dss.GetRepoId(),
		PersistentIndex: dss.GetIndex().IsPersistent(),
		Encrypted:       dss.IsRepoEncrypted(),
		ClientIsKnown:   cik,
	}
}

func aRecordClient(clId string, dss HDss) *mUpdatedData {
	ud, err := dss.GetIndex().recordClient(clId)
	if err != nil {
		return &mUpdatedData{mError: mError{Error: err.Error()}}
	}
	return &mUpdatedData{UpdatedData: ud}
}

func aUpdateClient(clId string, isFull bool, dss HDss) *mUpdatedData {
	ud, err := dss.GetIndex().updateClient(clId, isFull)
	if err != nil {
		return &mUpdatedData{mError: mError{Error: err.Error()}}
	}
	return &mUpdatedData{UpdatedData: ud}
}

func aQueryMetaTimes(npath string, dss HDss) *mTimes {
	ts, err := dss.(*ODss).proxy.queryMetaTimes(npath)
	if err != nil {
		return &mTimes{mError: mError{Error: err.Error()}}
	}
	return &mTimes{Times: ts}
}

func aStoreMeta(npath string, time int64, bs []byte, dss HDss) error {
	return dss.(*ODss).proxy.storeAndIndexMeta(npath, time, bs)
}

func aXStoreMeta(npath string, time int64, bs []byte, acl []ACLEntry, dss HDss) error {
	return dss.GetIndex().storeMeta(npath, time, bs)
}

func aRemoveMeta(npath string, time int64, dss HDss) error {
	return dss.(*ODss).proxy.removeMeta(npath, time)
}

func aXRemoveMeta(npath string, time int64, dss HDss) error {
	return dss.GetIndex().removeMeta(npath, time)
}

func aLoadMeta(npath string, time int64, dss HDss) *mLoadMetaOut {
	bs, err := dss.(*ODss).proxy.loadMeta(npath, time)
	if err != nil {
		return &mLoadMetaOut{mError: mError{Error: err.Error()}}
	}
	return &mLoadMetaOut{Bs: bs}
}

func aQueryContent(ch string, dss HDss) *mExist {
	ex, err := dss.(*ODss).proxy.queryContent(ch)
	if err != nil {
		return &mExist{mError: mError{Error: err.Error()}}
	}
	return &mExist{Exist: ex}
}

func cInitialize(apc WebApiClient) (*mInitialized, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mInitialized
	if wdc.LibApi {
		out = *aInitialize(wdc.ClId, wdc.libDss)
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"initialize/"+wdc.ClId, nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cInitialize: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cInitialize: %s", out.Error)
	}
	return &out, nil
}

func cRecordClient(apc WebApiClient) (*mUpdatedData, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mUpdatedData
	if wdc.LibApi {
		out = *aRecordClient(wdc.ClId, wdc.libDss)
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodPut, apc.Url()+"recordClient/"+wdc.ClId, nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cRecordClient: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cRecordClient: %s", out.Error)
	}
	return &out, nil
}

func cUpdateClient(apc WebApiClient, isFull bool) (*mUpdatedData, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mUpdatedData
	if wdc.LibApi {
		out = *aUpdateClient(wdc.ClId, isFull, wdc.libDss)
	} else {
		pif := ""
		if isFull {
			pif = "&isFull=true"
		}
		_, err := apc.SimpleDoAsJson(http.MethodPut, apc.Url()+"updateClient/"+wdc.ClId+pif, nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cUpdateClient: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cUpdateClient: %s", out.Error)
	}
	return &out, nil
}

func cLoadMeta(apc WebApiClient, npath string, time int64) (*mLoadMetaOut, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mLoadMetaOut
	if wdc.LibApi {
		out = *aLoadMeta(npath, time, wdc.libDss)
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"loadMeta", mLoadMetaIn{Npath: npath, Time: time}, &out)
		if err != nil {
			return nil, fmt.Errorf("in cLoadMeta: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cLoadMeta: %s", out.Error)
	}
	return &out, nil
}

func cQueryMetaTimes(apc WebApiClient, npath string) (*mTimes, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mTimes
	if wdc.LibApi {
		out = *aQueryMetaTimes(npath, wdc.libDss)
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"queryMetaTimes", npath, &out)
		if err != nil {
			return nil, fmt.Errorf("in cQueryMetaTimes: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cQueryMetaTimes: %s", out.Error)
	}
	return &out, nil
}

func cStoreMeta(apc WebApiClient, npath string, time int64, bs []byte) error {
	wdc := apc.GetConfig().(webDssClientConfig)
	var err error
	if wdc.LibApi {
		err = aStoreMeta(npath, time, bs, wdc.libDss)
	} else {
		_, err = apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"storeMeta", mStoreMeta{Npath: npath, Time: time, Bs: bs}, nil)
	}
	if err != nil {
		return fmt.Errorf("in cStoreMeta: %v", err)
	}
	return nil
}

func cRemoveMeta(apc WebApiClient, npath string, time int64) error {
	wdc := apc.GetConfig().(webDssClientConfig)
	var err error
	if wdc.LibApi {
		err = aRemoveMeta(npath, time, wdc.libDss)
	} else {
		_, err = apc.SimpleDoAsJson(http.MethodDelete, apc.Url()+"removeMeta", mRemoveMeta{Npath: npath, Time: time}, nil)
	}
	if err != nil {
		return fmt.Errorf("in cRemoveMeta: %v", err)
	}
	return nil
}

func cXRemoveMeta(apc WebApiClient, npath string, time int64) error {
	wdc := apc.GetConfig().(webDssClientConfig)
	var err error
	if wdc.LibApi {
		err = aXRemoveMeta(npath, time, wdc.libDss)
	} else {
		_, err = apc.SimpleDoAsJson(http.MethodDelete, apc.Url()+"xRemoveMeta", mRemoveMeta{Npath: npath, Time: time}, nil)
	}
	if err != nil {
		return fmt.Errorf("in cXRemoveMeta: %v", err)
	}
	return nil
}

func cQueryContent(apc WebApiClient, ch string) (*mExist, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mExist
	if wdc.LibApi {
		out = *aQueryContent(ch, wdc.libDss)
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"queryContent/"+ch, nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cQueryContent: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cQueryContent: %s", out.Error)
	}
	return &out, nil
}

func cDumpIndex(apc WebApiClient) (*mDump, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mDump
	if wdc.LibApi {
		out.Dump = wdc.libDss.DumpIndex()
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"dumpIndex", nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cDumpIndex: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cDumpIndex: %s", out.Error)
	}
	return &out, nil
}

func cScanPhysicalStorage(apc WebApiClient) (*mSPS, error) {
	wdc := apc.GetConfig().(webDssClientConfig)
	var out mSPS
	if wdc.LibApi {
		sti, errs := wdc.libDss.ScanStorage()
		if errs == nil {
			errs = &ErrorCollector{}
		}
		out.Sti, out.Errs = sti, *errs
	} else {
		_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"scanPhysicalStorage", nil, &out)
		if err != nil {
			return nil, fmt.Errorf("in cScanPhysicalStorage: %v", err)
		}
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cScanPhysicalStorage: %s", out.Error)
	}
	return &out, nil
}

func cOpenSession(apc WebApiClient, aclusers []string) error {
	wdc := apc.GetConfig().(webDssClientConfig)
	if len(aclusers) == 0 {
		aclusers = []string{""}
	}
	for _, au := range aclusers {
		var (
			i  int
			id IdentityConfig
		)
		for i, id = range wdc.identities {
			if id.Alias != au {
				continue
			}
		}
		if i == len(wdc.identities) {
			return fmt.Errorf("in cOpenSession: acluser \"%s\" not found as identity alias", au)
		}
		em, err := EncryptMsg("", id.PKey)
		if err != nil {
			return fmt.Errorf("in cOpenSession: acluser \"%s\" %w", au, err)
		}
		if id.Secret == "" {
			continue
		}
		if dm, err := DecryptMsg(em, id.Secret); err != nil || dm != "" {
			return fmt.Errorf("in cOpenSession: acluser \"%s\" %w \"%s\"", au, err, dm)
		}
	}
	return nil
}
