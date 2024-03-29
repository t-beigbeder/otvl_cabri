package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
	"strings"
)

type WebDssServerConfig struct {
	WebServerConfig
	UserConfig
	Dss HDss
}

func sInitialize(c echo.Context) error {
	clId := ""
	if err := echo.PathParamsBinder(c).String("clId", &clId).BindError(); err != nil {
		return NewServerErr("sInitialize", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aInitialize(clId, dss))
}

func sRecordClient(c echo.Context) error {
	clId := ""
	if err := echo.PathParamsBinder(c).String("clId", &clId).BindError(); err != nil {
		return NewServerErr("sRecordClient", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusCreated, aRecordClient(clId, dss))
}

func sUpdateClient(c echo.Context) error {
	clId := ""
	if err := echo.PathParamsBinder(c).String("clId", &clId).BindError(); err != nil {
		return NewServerErr("sUpdateClient", err)
	}
	isFull := false
	if err := echo.QueryParamsBinder(c).Bool("isFull", &isFull).BindError(); err != nil {
		return NewServerErr("sUpdateClient", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aUpdateClient(clId, isFull, dss))
}

func sQueryMetaTimes(c echo.Context) error {
	npath := ""
	if err := c.Bind(&npath); err != nil {
		return NewServerErr("sQueryMetaTimes", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aQueryMetaTimes(npath, dss))
}

func sStoreMeta(c echo.Context) error {
	var sm mStoreMeta
	if err := c.Bind(&sm); err != nil {
		return NewServerErr("sStoreMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := aStoreMeta(sm.Npath, sm.Time, sm.Bs, dss)
	if err != nil {
		return NewServerErr("sStoreMeta", err)
	}
	return c.JSON(http.StatusOK, nil)
}

func sRemoveMeta(c echo.Context) error {
	var rm mRemoveMeta
	if err := c.Bind(&rm); err != nil {
		return NewServerErr("sRemoveMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := aRemoveMeta(rm.Npath, rm.Time, dss)
	if err != nil {
		return NewServerErr("sRemoveMeta", err)
	}
	return c.JSON(http.StatusOK, nil)
}

func sXRemoveMeta(c echo.Context) error {
	var rm mRemoveMeta
	if err := c.Bind(&rm); err != nil {
		return NewServerErr("sXRemoveMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := aXRemoveMeta(rm.Npath, rm.Time, dss)
	if err != nil {
		return NewServerErr("sXRemoveMeta", err)
	}
	return c.JSON(http.StatusOK, nil)
}

func sPushContent(c echo.Context) error {
	req := c.Request()
	slja := make([]byte, 16)
	if n, err := req.Body.Read(slja); n != 16 || err != nil {
		return NewServerErr("sPushContent", fmt.Errorf("%d %v", n, err))
	}
	lja, err := internal.Str16ToInt64(string(slja))
	if err != nil {
		return NewServerErr("sPushContent", err)
	}
	jsonArgs := make([]byte, lja)
	if n, err := req.Body.Read(jsonArgs); n != len(jsonArgs) || (err != nil && err != io.EOF) {
		return NewServerErr("sPushContent", fmt.Errorf("%d %v", n, err))
	}
	args := mPushContentIn{}
	err = json.Unmarshal(jsonArgs, &args)
	if err != nil {
		return NewServerErr("sPushContent", err)
	}
	oDss := GetCustomConfig(c).(WebDssServerConfig).Dss.(*ODss)
	wter, err := oDss.proxy.spGetContentWriter(contentWriterCbs{
		getMetaBytes: func(iErr error, size int64, ch string) (mbs []byte, emid string, oErr error) {
			return args.Mbs, args.Emid, nil
		},
	}, nil)
	if err != nil {
		return NewServerErr("sPushContent", err)
	}
	n, err := io.Copy(wter, req.Body)
	if err != nil || n != args.Size {
		return NewServerErr("sPushContent", fmt.Errorf("%v %d %d", err, n, args.Size))
	}
	if err = wter.Close(); err != nil {
		return NewServerErr("sPushContent", err)
	}
	return c.JSON(http.StatusOK, &mError{})
}

func sLoadMeta(c echo.Context) error {
	var lm mLoadMetaIn
	if err := c.Bind(&lm); err != nil {
		return NewServerErr("sLoadMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aLoadMeta(lm.Npath, lm.Time, dss))
}

func sSpGetContentReader(c echo.Context) error {
	var args mSpGetContentReader
	if err := c.Bind(&args); err != nil {
		return NewServerErr("sDoGetContentReader", err)
	}
	oDss := GetCustomConfig(c).(WebDssServerConfig).Dss.(*ODss)
	resp := c.Response()
	resp.Writer.Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	rder, err := oDss.proxy.spGetContentReader(args.Ch)
	if err != nil {
		resp.WriteHeader(http.StatusOK)
		sErr := err.Error()
		io.Copy(resp.Writer, strings.NewReader(internal.Int64ToStr16(int64(len(sErr)))))
		io.Copy(resp.Writer, strings.NewReader(sErr))
		return nil
	}
	defer rder.Close()
	resp.WriteHeader(http.StatusOK)
	io.Copy(resp.Writer, strings.NewReader(internal.Int64ToStr16(int64(0))))
	io.Copy(resp.Writer, rder)
	return nil
}

func sQueryContent(c echo.Context) error {
	ch := ""
	if err := echo.PathParamsBinder(c).String("ch", &ch).BindError(); err != nil {
		return NewServerErr("sQueryContent", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aQueryContent(ch, dss))
}

func sRemoveContent(c echo.Context) error {
	ch := ""
	if err := echo.PathParamsBinder(c).String("ch", &ch).BindError(); err != nil {
		return NewServerErr("sRemoveContent", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := aRemoveContent(ch, dss)
	if err != nil {
		return NewServerErr("sRemoveContent", err)
	}
	return c.JSON(http.StatusOK, nil)
}

func sDumpIndex(c echo.Context) error {
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, &mDump{Dump: dss.DumpIndex()})
}

func sScanPhysicalStorage(c echo.Context) error {
	var checksum bool
	if err := echo.QueryParamsBinder(c).Bool("checksum", &checksum).BindError(); err != nil {
		return NewServerErr("sScanPhysicalStorage", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	sti, errs := dss.ScanStorage(checksum, false, false)
	if errs == nil {
		errs = &ErrorCollector{}
	}
	return c.JSON(http.StatusOK, &mSPS{Sti: sti, Errs: *errs})
}

func sLoadIndex(c echo.Context) error {
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	_, metas, _, err := dss.GetIndex().(*pIndex).loadInMemory()
	if err != nil {
		return NewServerErr("sLoadIndex", err)
	}
	return c.JSON(http.StatusOK, &mLoadedIndex{Metas: metas})
}

func WebDssServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
	dss := configs[root].(WebDssServerConfig).Dss
	_ = dss
	e.GET(root+"initialize/:clId", sInitialize)
	e.PUT(root+"recordClient/:clId", sRecordClient)
	e.PUT(root+"updateClient/:clId", sUpdateClient)
	e.POST(root+"queryMetaTimes", sQueryMetaTimes)
	e.POST(root+"storeMeta", sStoreMeta)
	e.DELETE(root+"removeMeta", sRemoveMeta)
	e.DELETE(root+"xRemoveMeta", sXRemoveMeta)
	e.POST(root+"pushContent", sPushContent)
	e.POST(root+"loadMeta", sLoadMeta)
	e.POST(root+"spGetContentReader", sSpGetContentReader)
	e.GET(root+"queryContent/:ch", sQueryContent)
	e.DELETE(root+"removeContent/:ch", sRemoveContent)
	e.GET(root+"dumpIndex", sDumpIndex)
	e.GET(root+"scanPhysicalStorage", sScanPhysicalStorage)
	e.GET(root+"loadIndex", sLoadIndex)
	return nil
}

func NewWebDssServer(root string, config WebDssServerConfig) (WebServer, error) {
	var tlsConfig *TlsConfig
	if config.IsTls {
		tlsConfig = getTlsServerConfig(config.WebServerConfig)
	}
	s := NewEServer(config.Addr, config.HasLog, tlsConfig)
	s.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
		return customConfigs[root].(WebDssServerConfig).Dss.Close()
	},
		WebDssServerConfigurator)
	err := s.Serve()
	return s, err
}
