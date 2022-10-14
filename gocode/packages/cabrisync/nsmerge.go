package cabrisync

import "github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"

func (syc *syncCtx) evalNsMerge() {
	syc.leftAndRight = append(syc.leftAndRight, syc.left.exCh...)
	syc.leftMg = append(syc.leftMg, syc.left.exCh...)
	syc.leftRight = append(syc.leftRight, syc.left.exCh...)
	syc.rightMg = append(syc.rightMg, syc.right.exCh...)
	for _, lch := range syc.left.exCh {
		if !internal.NpType(lch).ExistIn(syc.right.exCh) {
			syc.rightMg = append(syc.rightMg, lch)
		}
	}
	for _, rch := range syc.right.exCh {
		if !internal.NpType(rch).ExistIn(syc.left.exCh) {
			syc.leftAndRight = append(syc.leftAndRight, rch)
			if !syc.options.KeepContent {
				if syc.options.BiDir {
					syc.leftMg = append(syc.leftMg, rch)
					syc.leftRight = append(syc.leftRight, rch)
				} else {
					syc.rmRight = append(syc.rmRight, rch)
				}
			} else {
				syc.leftRight = append(syc.leftRight, rch)
			}
		}
	}
	return
}
