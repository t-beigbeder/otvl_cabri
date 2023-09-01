package cabridss

type wfsDss struct {
	apc    WebApiClient
	closed bool
}

type WfsDssConfig struct {
	DssBaseConfig
	NoClientLimit bool
}

// NewWfsDss opens a web client for a "fsy" DSS (data storage system)
// config provides the web client configuration
// returns a pointer to the ready to use DSS or an error if any occur
func NewWfsDss(config WfsDssConfig, slsttime int64, aclusers []string) (Dss, error) {
	panic("to be implemented NewWfsDss")
}
