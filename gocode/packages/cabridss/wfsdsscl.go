package cabridss

import (
	"fmt"
	"net/http"
)

type wfsDssClientConfig struct {
	DssBaseConfig
	NoClientLimit bool
}

func cfsInitialize(apc WebApiClient) (*mError, error) {
	//wdc := apc.GetConfig().(wfsDssClientConfig)
	var out mError
	_, err := apc.SimpleDoAsJson(http.MethodGet, apc.Url()+"wfsInitialize", nil, &out)
	if err != nil {
		return nil, fmt.Errorf("in cfsInitialize: %v", err)
	}
	if out.Error != "" {
		return nil, fmt.Errorf("in cfsInitialize: %s", out.Error)
	}
	return &out, nil
}
