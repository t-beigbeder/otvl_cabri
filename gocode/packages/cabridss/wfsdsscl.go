package cabridss

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
	"net/url"
)

type mfsMkupdateNs struct {
	Npath    string     `json:"npath"`
	Mtime    int64      `json:"mtime,string"`
	Children []string   `json:"children"`
	ACL      []ACLEntry `json:"acl"`
}

type mfsLsnsOut struct {
	mError
	Children []string `json:"children"`
}

type mfsGetContentWriterIn struct {
	Npath string     `json:"npath"`
	Mtime int64      `json:"mtime,string"`
	ACL   []ACLEntry `json:"acl"`
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
	var lo *mfsLsnsOut
	epath := url.PathEscape(npath)
	_, err = apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"wfsLsns/"+epath, nil, &lo)
	if err != nil {
		err = fmt.Errorf("in cfsLsns: %w", err)
		return
	}
	if lo.mError.Error != "" {
		err = fmt.Errorf("in cfsLsns: %s", lo.mError.Error)
	}
	children = lo.Children
	return
}

func cfsGetContentWriter(apc WebApiClient, npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (wc io.WriteCloser, err error) {
	jsonArgs, err := json.Marshal(mfsGetContentWriterIn{Npath: npath, Mtime: mtime, ACL: acl})
	if err != nil {
		return
	}
	lja := internal.Int64ToStr16(int64(len(jsonArgs)))
	_ = lja
	pr, wc := io.Pipe()
	hdler := webContentWriterHandler{header: make([]byte, 16+len(jsonArgs)), rCloser: pr, hasHash: true}
	copy(hdler.header, lja)
	copy(hdler.header[16:], jsonArgs)
	go func() {
		var err error
		defer func() {
			if cb != nil {
				if err != nil {
					cb(err, 0, "")
				} else {
					cb(nil, hdler.written, internal.Sha256ToStr32(hdler.h.Sum(nil)))
				}
			}
			if err != nil && cb != nil {
				cb(err, 0, "")
			}
			if err != nil {
				err = fmt.Errorf("in cfsGetContentWriter cb: %w", err)
			}
		}()
		err = errors.New("go func to be tested")
		req, err := http.NewRequest(http.MethodPost, apc.Url()+"wfsGetContentWriter", nil)
		req.Body = &hdler
		req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
		resp, err := apc.(*apiClient).client.Do(req)
		if err = NewClientErr("", resp, err, nil); err != nil {
			return
		}
		bs, err := io.ReadAll(resp.Body)
		var pco mError
		if err = json.Unmarshal(bs, &pco); err != nil {
			return
		}
		if pco.Error != "" {
			err = errors.New(pco.Error)
			return
		}
	}()
	return
}
