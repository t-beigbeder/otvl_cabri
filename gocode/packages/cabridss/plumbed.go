package cabridss

import "github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"

func CabriPlumberDefaultConfig(serial bool, redLimit int) plumber.Config {
	gm := uint(8)
	sgm := uint(12)
	return plumber.Config{
		PlizerEnabled: !serial,
		RglatorsByName: map[string]uint{
			"LsnsMetas": 0, "GetMetas": gm,
			"Synchronize": 0, "SyncEntries": 0, "SyncGetMetas": sgm,
		},
	}
}
