package cabridss

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"net/http"
	"net/url"
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
	if n, err := req.Body.Read(jsonArgs); n != len(jsonArgs) || err != nil {
		return NewServerErr("sfsGetContentWriter", fmt.Errorf("%d %v", n, err))
	}
	args := mfsGetContentWriterIn{}
	err = json.Unmarshal(jsonArgs, &args)
	if err != nil {
		return NewServerErr("sfsGetContentWriter", err)
	}
	if err != nil {
		return NewServerErr("sfsGetContentWriter", err)
	}
	//dss := GetCustomConfig(c).(WfsDssServerConfig).Dss
	return NewServerErr("sfsGetContentWriter", fmt.Errorf("to be implemented"))
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
