package cabridss

import (
	"github.com/labstack/echo/v4"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"io"
	"net/http"
)

func sRestGet(c echo.Context) error {
	req := c.Request()
	if req.Header.Get("Cabri") == "WebApi" {
		// hack, as remote API starts with GET <path>/initialize/xxx
		// this will provide an explicit error to the DSS client
		return c.JSON(http.StatusOK,
			&mInitialized{mError: mError{Error: "REST API here, not a remote DSS Web API!"}})
	}
	path := ""
	if err := echo.PathParamsBinder(c).String("path", &path).BindError(); err != nil {
		return NewServerErr("sRestGet", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	qm, ok := c.QueryParams()["meta"]
	_, _ = qm, ok
	var im IMeta
	var err error
	im, err = dss.GetMeta(path, true)
	if err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	if ok {
		return c.JSON(http.StatusOK, im)
	}
	if im.GetIsNs() {
		return c.JSON(http.StatusOK, im.GetChildren())
	}
	resp := c.Response()
	resp.Writer.Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	rder, err := dss.GetContentReader(path)
	if err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	defer rder.Close()
	resp.WriteHeader(http.StatusOK)
	io.Copy(resp.Writer, rder)
	return nil
}

func getUpdateQueryParams(c echo.Context) (mtime int64, acl []ACLEntry, err error) {
	var smtime string
	var sacl []string
	if err = echo.QueryParamsBinder(c).String("mtime", &smtime).BindError(); err != nil {
		return
	}
	if mtime, err = internal.CheckTimeStamp(smtime); err != nil {
		err = &ErrBadParameter{Key: "mtime", Value: internal.StringStringer(smtime), Err: err}
		return
	}
	if err = echo.QueryParamsBinder(c).Strings("acl", &sacl).BindError(); err != nil {
		return
	}
	if acl, err = CheckUiACL(sacl); err != nil {
		err = &ErrBadParameter{Key: "acl", Value: internal.StringsStringer(sacl), Err: err}
		return
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	uc := GetCustomConfig(c).(WebDssServerConfig).UserConfig
	if !dss.IsEncrypted() {
		return
	}
	var acl2 []ACLEntry
	for _, uac := range acl {
		if idc := uc.GetIdentity(uac.User); idc.PKey != "" {
			acl2 = append(acl2, ACLEntry{User: idc.PKey, Rights: uac.Rights})
		} else {
			acl2 = append(acl2, uac)
		}
	}
	acl = acl2
	return
}

func sRestPost(c echo.Context) error {
	path := ""
	if err := echo.PathParamsBinder(c).String("path", &path).BindError(); err != nil {
		return NewServerErr("sRestPost", err)
	}
	mtime, acl, err := getUpdateQueryParams(c)
	if err != nil {
		_, bpe := err.(*ErrBadParameter)
		if bpe {
			return c.JSON(http.StatusUnprocessableEntity, &mError{Error: err.Error()})
		}
		return err
	}
	var children []string
	if err = echo.QueryParamsBinder(c).Strings("child", &children).BindError(); err != nil {
		return err
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	if err := dss.Updatens(path, mtime, children, acl); err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	return c.NoContent(http.StatusCreated)
}

func sRestPut(c echo.Context) error {
	path := ""
	if err := echo.PathParamsBinder(c).String("path", &path).BindError(); err != nil {
		return NewServerErr("sRestPut", err)
	}
	mtime, acl, err := getUpdateQueryParams(c)
	if err != nil {
		_, bpe := err.(*ErrBadParameter)
		if bpe {
			return c.JSON(http.StatusUnprocessableEntity, &mError{Error: err.Error()})
		}
		return err
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	wter, err := dss.GetContentWriter(path, mtime, acl, nil)
	if err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	req := c.Request()
	_, err = io.Copy(wter, req.Body)
	if err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	if err = wter.Close(); err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	return c.NoContent(http.StatusCreated)
}

func sRestDelete(c echo.Context) error {
	path := ""
	if err := echo.PathParamsBinder(c).String("path", &path).BindError(); err != nil {
		return NewServerErr("sRestDelete", err)
	}
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	if err := dss.Remove(path); err != nil {
		return c.JSON(http.StatusConflict, &mError{Error: err.Error()})
	}
	return c.NoContent(http.StatusOK)
}

func RestServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
	e.GET(root, sRestGet)
	e.GET(root+":path", sRestGet)
	e.POST(root, sRestPost)
	e.POST(root+":path", sRestPost)
	e.PUT(root+":path", sRestPut)
	e.DELETE(root, sRestDelete)
	e.DELETE(root+":path", sRestDelete)
	return nil
}

func NewRestServer(root string, config WebDssServerConfig) (WebServer, error) {
	var tlsConfig *TlsConfig
	if config.IsTls {
		tlsConfig = getTlsServerConfig(config.WebServerConfig)
	}
	s := NewEServer(config.Addr, config.HasLog, tlsConfig)
	s.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
		return customConfigs[root].(WebDssServerConfig).Dss.Close()
	},
		RestServerConfigurator)
	err := s.Serve()
	return s, err
}
