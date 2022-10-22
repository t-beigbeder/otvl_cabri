package cabridss

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
	"os"
	"strings"
)

type WebDssServerConfig struct {
	Dss              HDss
	HasLog           bool
	ShutdownCallback func(err error) error
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

func sXStoreMeta(c echo.Context) error {
	var sm mStoreMeta
	if err := c.Bind(&sm); err != nil {
		return NewServerErr("sXStoreMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := aXStoreMeta(sm.Npath, sm.Time, sm.Bs, dss)
	if err != nil {
		return NewServerErr("sXStoreMeta", err)
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

func sOnCloseContent(c echo.Context) error {
	req := c.Request()
	slja := make([]byte, 16)
	if n, err := req.Body.Read(slja); n != 16 || err != nil {
		return NewServerErr("sStoreMeta", fmt.Errorf("%d %v", n, err))
	}
	lja, err := internal.Str16ToInt64(string(slja))
	if err != nil {
		return NewServerErr("sStoreMeta", err)
	}
	jsonArgs := make([]byte, lja)
	if n, err := req.Body.Read(jsonArgs); n != len(jsonArgs) || err != nil {
		return NewServerErr("sStoreMeta", fmt.Errorf("%d %v", n, err))
	}
	args := mOnCloseContentIn{}
	err = json.Unmarshal(jsonArgs, &args)
	if err != nil {
		return NewServerErr("sStoreMeta", err)
	}

	cf, err := os.CreateTemp("", "scw")
	if err != nil {
		return NewServerErr("sStoreMeta", err)
	}
	oDss := GetCustomConfig(c).(WebDssServerConfig).Dss.(*ODss)
	var cbErr error
	var cbOut mOnCloseContentOut
	lcb := func(err error, size int64, sha256trunc []byte) {
		if err == nil {
			cbErr = oDss.proxy.onCloseContent(args.Npath, args.Mtime, cf, size, sha256trunc, args.ACL, func(npath string, time int64, bs []byte) error {
				cbOut = mOnCloseContentOut{Npath: npath, Time: time, Bs: bs}
				if err = oDss.proxy.xStoreMeta(npath, time, bs); err != nil {
					return fmt.Errorf("in onCloseContent: %w", err)
				}
				return oDss.proxy.storeMeta(npath, time, bs)
			})
		}
	}
	wter := &ContentHandle{cb: lcb, cf: cf, h: sha256.New()}
	n, err := io.Copy(wter, req.Body)
	if err != nil || n != args.Size {
		return NewServerErr("sStoreMeta", fmt.Errorf("%v %d %d", err, n, args.Size))
	}
	if err = wter.Close(); err != nil {
		return NewServerErr("sStoreMeta", err)
	}
	if cbErr != nil {
		return NewServerErr("sStoreMeta", cbErr)
	}
	return c.JSON(http.StatusOK, &cbOut)
}

func sLoadMeta(c echo.Context) error {
	var lm mLoadMetaIn
	if err := c.Bind(&lm); err != nil {
		return NewServerErr("sLoadMeta", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, aLoadMeta(lm.Npath, lm.Time, dss))
}

func sDoGetContentReader(c echo.Context) error {
	var args mDoGetContentReader
	if err := c.Bind(&args); err != nil {
		return NewServerErr("sDoGetContentReader", err)
	}
	oDss := GetCustomConfig(c).(WebDssServerConfig).Dss.(*ODss)
	resp := c.Response()
	resp.Writer.Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	rder, err := oDss.proxy.doGetContentReader(args.Npath, args.MData)
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

func sDumpIndex(c echo.Context) error {
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	return c.JSON(http.StatusOK, &mDump{Dump: dss.DumpIndex()})
}

func sScanPhysicalStorage(c echo.Context) error {
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	sti, errs := dss.ScanStorage()
	if errs == nil {
		errs = &ErrorCollector{}
	}
	return c.JSON(http.StatusOK, &mSPS{Sti: sti, Errs: *errs})
}

func WebDssServerConfigurator(e *echo.Echo, root string, config interface{}) error {
	dss := config.(WebDssServerConfig).Dss
	_ = dss
	e.GET(root+"initialize/:clId", sInitialize)
	e.PUT(root+"recordClient/:clId", sRecordClient)
	e.PUT(root+"updateClient/:clId", sUpdateClient)
	e.POST(root+"queryMetaTimes", sQueryMetaTimes)
	e.POST(root+"storeMeta", sStoreMeta)
	e.POST(root+"xStoreMeta", sXStoreMeta)
	e.DELETE(root+"removeMeta", sRemoveMeta)
	e.DELETE(root+"xRemoveMeta", sXRemoveMeta)
	e.POST(root+"onCloseContent", sOnCloseContent)
	e.POST(root+"loadMeta", sLoadMeta)
	e.POST(root+"doGetContentReader", sDoGetContentReader)
	e.GET(root+"queryContent/:ch", sQueryContent)
	e.GET(root+"dumpIndex", sDumpIndex)
	e.GET(root+"scanPhysicalStorage", sScanPhysicalStorage)
	return nil
}

func NewWebDssServer(addr, root string, config WebDssServerConfig) (WebServer, error) {
	s := NewEServer(addr, config.HasLog)
	s.ConfigureApi(root, config, WebDssServerConfigurator, func(customConfig interface{}) error {
		err := config.Dss.Close()
		if config.ShutdownCallback != nil {
			err = config.ShutdownCallback(err)
		}
		return err
	})
	err := s.Serve()
	return s, err
}
