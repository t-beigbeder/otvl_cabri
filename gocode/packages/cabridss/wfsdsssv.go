package cabridss

import "github.com/labstack/echo/v4"

func WfsDssServerConfigurator(e *echo.Echo, root string, configs map[string]interface{}) error {
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
