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

type mfsGetMetaOut struct {
	mError
	MetaOut Meta `json:"meta"`
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

type cfsClientWriter struct {
	wwsc       chan struct{}
	buffer     []byte
	xBuf       int
	wrsc       chan struct{}
	readerErr  error
	readerDone bool
	info       interface{}
}

func newCfsClientWriter(info interface{}) *cfsClientWriter {
	return &cfsClientWriter{wwsc: make(chan struct{}), wrsc: make(chan struct{}), info: info}
}

func (ccw *cfsClientWriter) Write(p []byte) (n int, err error) {
	<-ccw.wrsc
	if ccw.readerErr != nil {
		return 0, fmt.Errorf("in cfsClientWriter.Write: reader error %w", ccw.readerErr)
	}
	ccw.buffer = make([]byte, len(p))
	ccw.xBuf = 0
	copy(ccw.buffer, p)
	ccw.wwsc <- struct{}{}
	return len(p), nil
}

func (ccw *cfsClientWriter) Close() error {
	close(ccw.wwsc)
	for !ccw.readerDone {
		<-ccw.wrsc
	}
	ccw.buffer = nil
	ccw.xBuf = 0
	return ccw.readerErr
}

type cfsClientWriterReader struct {
	ccw    *cfsClientWriter
	header []byte
	offset int
}

func (ccwr *cfsClientWriterReader) Read(p []byte) (n int, err error) {
	if ccwr.offset < len(ccwr.header) {
		l := len(ccwr.header) - ccwr.offset
		if l > len(p) {
			l = len(p)
		}
		copy(p, ccwr.header[ccwr.offset:ccwr.offset+l])
		ccwr.offset += l
		n += l
	}
	if n == len(p) {
		return
	}
	if ccwr.ccw.xBuf == 0 {
		<-ccwr.ccw.wwsc
	}
	l := len(ccwr.ccw.buffer) - ccwr.ccw.xBuf

	if n+l > len(p) {
		l = len(p) - n
	}
	if l == 0 {
		err = io.EOF
		return
	}
	copy(p[n:n+l], ccwr.ccw.buffer[ccwr.ccw.xBuf:ccwr.ccw.xBuf+l])
	ccwr.ccw.xBuf += l
	n += l
	ccwr.offset += l
	if ccwr.ccw.xBuf == len(ccwr.ccw.buffer) && len(ccwr.ccw.buffer) > 0 {
		ccwr.ccw.buffer = nil
		ccwr.ccw.xBuf = 0
		ccwr.ccw.wrsc <- struct{}{}
	}
	return
}

func (ccwr *cfsClientWriterReader) Close() error {
	return nil
}

func (ccwr *cfsClientWriterReader) Done(err error) {
	select {
	case <-ccwr.ccw.wwsc:
	default:
	}
	ccwr.ccw.readerErr = err
	ccwr.ccw.readerDone = true
	close(ccwr.ccw.wrsc)
}

func newCfsClientWriterReader(ccw *cfsClientWriter, header []byte) *cfsClientWriterReader {
	ccwr := &cfsClientWriterReader{ccw: ccw, header: header}
	ccwr.ccw.wrsc <- struct{}{}
	return ccwr
}

func cfsGetContentWriter(apc WebApiClient, npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (pcw io.WriteCloser, err error) {
	jsonArgs, err := json.Marshal(mfsGetContentWriterIn{Npath: npath, Mtime: mtime, ACL: acl})
	if err != nil {
		return
	}
	ccw := newCfsClientWriter(npath)
	go func() {
		var (
			err  error
			req  *http.Request
			resp *http.Response
			bs   []byte
		)
		req, err = http.NewRequest(http.MethodPost, apc.Url()+"wfsGetContentWriter", nil)
		lja := internal.Int64ToStr16(int64(len(jsonArgs)))
		header := make([]byte, 16+len(jsonArgs))
		copy(header, lja)
		copy(header[16:], jsonArgs)
		ccwr := newCfsClientWriterReader(ccw, header)
		defer func() {
			ccwr.Done(err)
		}()
		req.Body = ccwr
		req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
		resp, err = apc.(*apiClient).client.Do(req, nil)
		if err = NewClientErr("", resp, err, nil); err != nil {
			return
		}
		bs, err = io.ReadAll(resp.Body)
		var pco mError
		if err = json.Unmarshal(bs, &pco); err != nil {
			return
		}
		if pco.Error != "" {
			err = errors.New(pco.Error)
			return
		}
		return
	}()
	return ccw, nil
}

func cfsGetContentReader(apc WebApiClient, npath string) (io.ReadCloser, error) {
	epath := url.PathEscape(npath)
	req, err := http.NewRequest(http.MethodGet, apc.Url()+"wfsGetContentReader/"+epath, nil)
	if err != nil {
		return nil, fmt.Errorf("in cfsGetContentReader: %w", err)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	resp, err := apc.(*apiClient).client.Do(req, nil)
	if err != nil {
		return nil, fmt.Errorf("in cfsGetContentReader: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		bs, err := io.ReadAll(resp.Body)
		return nil, NewClientErr("cfsGetContentReader", resp, err, bs)
	}
	slj := make([]byte, 16)
	if n, err := resp.Body.Read(slj); n != 16 || (err != nil && err != io.EOF) {
		return nil, fmt.Errorf("in cfsGetContentReader: %w", err)
	}
	lj, err := internal.Str16ToInt64(string(slj))
	if err != nil {
		return nil, fmt.Errorf("in cfsGetContentReader: %w", err)
	}
	if lj != 0 {
		sErr := make([]byte, lj)
		if n, err := resp.Body.Read(sErr); n != int(lj) || (err != nil && err != io.EOF) {
			return nil, fmt.Errorf("in cfsGetContentReader: %w", err)
		}
		return nil, fmt.Errorf("in cfsGetContentReader: %s", sErr)
	}
	return resp.Body, nil
}

func cfsRemove(apc WebApiClient, npath string) (err error) {
	var rer mError
	epath := url.PathEscape(npath)
	_, err = apc.SimpleDoAsJson(http.MethodDelete, apc.Url()+"wfsRemove/"+epath, nil, &rer)
	if err != nil {
		return fmt.Errorf("in cfsRemove: %w", err)
	}
	if rer.Error != "" {
		return fmt.Errorf("in cfsRemove: %s", rer.Error)
	}
	return
}

func cfsGetMeta(apc WebApiClient, npath string, getCh bool) (meta IMeta, err error) {
	var gmo *mfsGetMetaOut
	uPath := fmt.Sprintf("wfsGetMeta/%s", url.PathEscape(npath))
	if getCh {
		uPath += "?getCh=true"
	}
	_, err = apc.SimpleDoAsJson(http.MethodGet, apc.Url()+uPath, nil, &gmo)
	if err != nil {
		return nil, fmt.Errorf("in cfsGetMeta: %w", err)
	}
	if gmo.Error != "" {
		return nil, fmt.Errorf("in cfsGetMeta: %s", gmo.Error)
	}
	meta = gmo.MetaOut
	return
}

func cfsSuEnableWrite(apc WebApiClient, npath string) (err error) {
	var rer mError
	epath := url.PathEscape(npath)
	_, err = apc.SimpleDoAsJson(http.MethodPut, apc.Url()+"wfsSuEnableWrite/"+epath, nil, &rer)
	if err != nil {
		return fmt.Errorf("in cfsSuEnableWrite: %w", err)
	}
	if rer.Error != "" {
		return fmt.Errorf("in cfsSuEnableWrite: %s", rer.Error)
	}
	return
}
