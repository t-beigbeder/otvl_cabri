package cabridss

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func afsInitialize(dss Dss) error {
	return nil
}

func sfsInitialize(c echo.Context) error {
	dss := GetCustomConfig(c).(WebDssServerConfig).Dss
	err := afsInitialize(dss)
	if err != nil {
		return NewServerErr("sfsInitialize", err)
	}
	return c.JSON(http.StatusOK, nil)
}

func WfsDssServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
	dss := configs[root].(WebDssServerConfig).Dss
	_ = dss
	e.GET(root+"wfsInitialize", sfsInitialize)
	return nil
}

func NewWfsDssServer(root string, config WebDssServerConfig) (WebServer, error) {
	var tlsConfig *TlsConfig
	if config.IsTls {
		tlsConfig = getTlsServerConfig(config.WebServerConfig)
	}
	s := NewEServer(config.Addr, config.HasLog, tlsConfig)
	s.ConfigureApi(root, config, func(root string, customConfigs map[string]interface{}) error {
		return customConfigs[root].(WebDssServerConfig).Dss.Close()
	},
		WfsDssServerConfigurator)
	err := s.Serve()
	return s, err
}
