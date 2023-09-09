package cabridss

import (
	"fmt"
	"net/http"
)

type mfsMkupdateNs struct {
	Npath    string     `json:"npath"`
	Mtime    int64      `json:"mtime,string"`
	Children []string   `json:"children"`
	ACL      []ACLEntry `json:"acl"`
}

func cfsInitialize(apc WebApiClient) error {
	var out mError
	_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"wfsInitialize", nil, &out)
	if err != nil {
		return fmt.Errorf("in cfsInitialize: %v", err)
	}
	if out.Error != "" {
		return fmt.Errorf("in cfsInitialize: %s", out.Error)
	}
	return nil
}

func cfsMkupdateNs(apc WebApiClient, isUpd bool, npath string, mtime int64, children []string, acl []ACLEntry) error {
	//wdc := wdi.apc.GetConfig().(WfsDssConfig)
	var rer mError
	urlPath := "wfsMkns"
	if isUpd {
		urlPath = "wfsUpdatens"
	}
	_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+urlPath,
		mfsMkupdateNs{
			Npath:    npath,
			Mtime:    mtime,
			Children: children,
			ACL:      acl,
		}, &rer)
	if err != nil {
		return fmt.Errorf("in cfsMkns: %w", err)
	}
	if rer.Error != "" {
		return fmt.Errorf("in cfsMkns: %s", rer.Error)
	}
	return nil
}

func cfsMkns(apc WebApiClient, npath string, mtime int64, children []string, acl []ACLEntry) error {
	return cfsMkupdateNs(apc, false, npath, mtime, children, acl)
}

func cfsUpdatens(apc WebApiClient, npath string, mtime int64, children []string, acl []ACLEntry) error {
	return cfsMkupdateNs(apc, true, npath, mtime, children, acl)
}

func cfsLsns(apc WebApiClient, npath string) (children []string, err error) {
	return nil, nil
}
