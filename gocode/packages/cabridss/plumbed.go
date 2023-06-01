package cabridss

import "github.com/t-beigbeder/otvl_cabri/gocode/packages/plumber"

func CabriPlumberDefaultConfig(serial bool, redLimit int) plumber.Config {
	gm := uint(8)
	sgm := uint(12)
	//if redLimit != 0 {
	//	gm = uint(redLimit / 10)
	//	if gm < 1 {
	//		gm = 1
	//	}
	//	sgm = gm
	//}
	return plumber.Config{
		PlizerEnabled: !serial,
		RglatorsByName: map[string]uint{
			"LsnsMetas": 0, "GetMetas": gm,
			"Synchronize": 0, "SyncEntries": 0, "SyncGetMetas": sgm,
		},
	}
}
