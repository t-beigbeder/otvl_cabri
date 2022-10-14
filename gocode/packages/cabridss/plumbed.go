package cabridss

import "github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"

func CabriPlumberDefaultConfig(serial bool) plumber.Config {
	return plumber.Config{
		PlizerEnabled: !serial,
		RglatorsByName: map[string]uint{
			"LsnsMetas": 0, "GetMetas": 8,
			"Synchronize": 0, "SyncEntries": 0, "SyncGetMetas": 12,
		},
	}
}
