package cabridss

import (
	"fmt"
	"net/http"
)

type mfsUpdateNs struct {
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

func cfsUpdatens(apc WebApiClient, npath string, mtime int64, children []string, acl []ACLEntry) error {
	//wdc := wdi.apc.GetConfig().(WfsDssConfig)
	_, err := apc.SimpleDoAsJson(http.MethodPost, apc.Url()+"wfsUpdatens",
		mfsUpdateNs{
			Npath:    npath,
			Mtime:    mtime,
			Children: children,
			ACL:      acl,
		}, nil)
	return err
}
