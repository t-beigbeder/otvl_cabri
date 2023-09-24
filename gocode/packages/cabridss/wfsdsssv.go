package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type WfsDssServerConfig struct {
	WebServerConfig
	Dss Dss
}

func sfsInitialize(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func sfsMkns(c echo.Context) error {
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	var un mfsMkupdateNs
	if err := c.Bind(&un); err != nil {
		return NewServerErr("sfsUpdatens", err)
	}
	return c.JSON(http.StatusOK, err2mError(dss.Mkns(un.Npath, un.Mtime, un.Children, un.ACL)))
}

func sfsUpdatens(c echo.Context) error {
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	var un mfsMkupdateNs
	if err := c.Bind(&un); err != nil {
		return NewServerErr("sfsUpdatens", err)
	}
	return c.JSON(http.StatusOK, err2mError(dss.Updatens(un.Npath, un.Mtime, un.Children, un.ACL)))
}

func sfsLsnsWhatever(c echo.Context, npath string) error {
	npath, err := url.PathUnescape(npath)
	var lo mfsLsnsOut
	if err != nil {
		lo.Error = err.Error()
		return c.JSON(http.StatusOK, &lo)
	}
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	children, err := dss.Lsns(npath)
	lo.Children = children
	if err != nil {
		lo.Error = err.Error()
	}
	return c.JSON(http.StatusOK, &lo)
}

func sfsLsns(c echo.Context) error {
	npath := ""
	if err := echo.PathParamsBinder(c).String("npath", &npath).BindError(); err != nil {
		return NewServerErr("sfsLsns", err)
	}
	return sfsLsnsWhatever(c, npath)
}

func sfsLsnsRoot(c echo.Context) error {
	return sfsLsnsWhatever(c, "")
}

func sfsGetContentWriter(c echo.Context) error {
	req := c.Request()
	slja := make([]byte, 16)
	if n, err := req.Body.Read(slja); n != 16 || err != nil {
		return NewServerErr("sfsGetContentWriter", fmt.Errorf("%d %v", n, err))
	}
	lja, err := internal.Str16ToInt64(string(slja))
	if err != nil {
		return NewServerErr("sfsGetContentWriter", err)
	}
	jsonArgs := make([]byte, lja)
	if n, err := req.Body.Read(jsonArgs); n != len(jsonArgs) || (err != nil && err != io.EOF) {
		return NewServerErr("sfsGetContentWriter", fmt.Errorf("%d %v", n, err))
	}
	args := mfsGetContentWriterIn{}
	err = json.Unmarshal(jsonArgs, &args)
	if err != nil {
		return NewServerErr("sfsGetContentWriter", err)
	}
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	wc, err := dss.GetContentWriter(args.Npath, args.Mtime, args.ACL, nil)
	if err != nil {
		return c.JSON(http.StatusOK, &mError{Error: err.Error()})
	}
	defer wc.Close()
	_, err = io.Copy(wc, req.Body)
	if err != nil {
		return NewServerErr("sfsGetContentWriter", err)
	}
	return c.JSON(http.StatusOK, &mError{})
}

func sfsGetContentReader(c echo.Context) error {
	npath := ""
	if err := echo.PathParamsBinder(c).String("npath", &npath).BindError(); err != nil {
		return NewServerErr("sfsGetContentReader", err)
	}
	npath, err := url.PathUnescape(npath)
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	resp := c.Response()
	resp.Writer.Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	rc, err := dss.GetContentReader(npath)
	if err != nil {
		resp.WriteHeader(http.StatusOK)
		sErr := err.Error()
		io.Copy(resp.Writer, strings.NewReader(internal.Int64ToStr16(int64(len(sErr)))))
		io.Copy(resp.Writer, strings.NewReader(sErr))
		return nil
	}
	defer rc.Close()
	resp.WriteHeader(http.StatusOK)
	io.Copy(resp.Writer, strings.NewReader(internal.Int64ToStr16(int64(0))))
	io.Copy(resp.Writer, rc)
	return nil
}

func sfsRemove(c echo.Context) error {
	var (
		err   error
		npath string
	)
	if err = echo.PathParamsBinder(c).String("npath", &npath).BindError(); err != nil {
		return NewServerErr("sfsRemove", err)
	}
	npath, err = url.PathUnescape(npath)
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	return c.JSON(http.StatusOK, err2mError(dss.Remove(npath)))
}

func sfsGetMetaWhatever(c echo.Context, npath string) error {
	npath, err := url.PathUnescape(npath)
	var getCh bool
	if err = echo.QueryParamsBinder(c).Bool("getCh", &getCh).BindError(); err != nil {
		return NewServerErr("sfsGetMetaWhatever", err)
	}
	var gm mfsGetMetaOut
	if err != nil {
		gm.Error = err.Error()
		return c.JSON(http.StatusOK, &gm)
	}
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	mo, err := dss.GetMeta(npath, getCh)
	if mo != nil {
		gm.MetaOut = mo.(Meta)
	}
	if err != nil {
		gm.Error = err.Error()
	}
	return c.JSON(http.StatusOK, &gm)
}

func sfsGetMeta(c echo.Context) error {
	npath := ""
	if err := echo.PathParamsBinder(c).String("npath", &npath).BindError(); err != nil {
		return NewServerErr("sfsGetMeta", err)
	}
	return sfsGetMetaWhatever(c, npath)
}

func sfsGetMetaRoot(c echo.Context) error {
	return sfsGetMetaWhatever(c, "")
}

func sfsSuEnableWrite(c echo.Context) error {
	var (
		err   error
		npath string
	)
	if err = echo.PathParamsBinder(c).String("npath", &npath).BindError(); err != nil {
		return NewServerErr("sfsSuEnableWrite", err)
	}
	npath, err = url.PathUnescape(npath)
	dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	dss.SetSu() // FIXME: only if enabled by server
	return c.JSON(http.StatusOK, err2mError(dss.SuEnableWrite(npath)))
}

func WfsDssServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
	dss := configs[root].(WfsDssServerConfig).Dss
	_ = dss
	e.GET(root+"wfsInitialize", sfsInitialize)
	e.POST(root+"wfsMkns", sfsMkns)
	e.POST(root+"wfsUpdatens", sfsUpdatens)
	e.GET(root+"wfsLsns/:npath", sfsLsns)
	e.GET(root+"wfsLsns/", sfsLsnsRoot)
	e.POST(root+"wfsGetContentWriter", sfsGetContentWriter)
	e.GET(root+"wfsGetContentReader/:npath", sfsGetContentReader)
	e.DELETE(root+"wfsRemove/:npath", sfsRemove)
	e.GET(root+"wfsGetMeta/:npath", sfsGetMeta)
	e.GET(root+"wfsGetMeta/", sfsGetMetaRoot)
	e.PUT(root+"wfsSuEnableWrite/:npath", sfsSuEnableWrite)
	return nil
}

func NewWfsDssServer(root string, config WfsDssServerConfig) (WebServer, error) {
	var tlsConfig *TlsConfig
	if config.IsTls {
		tlsConfig = getTlsServerConfig(config.WebServerConfig)
	}
	s := NewEServer(config.Addr, config.HasLog, tlsConfig)
	s.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
		return customConfigs[root].(WfsDssServerConfig).Dss.Close()
	},
		WfsDssServerConfigurator)
	err := s.Serve()
	return s, err
}
