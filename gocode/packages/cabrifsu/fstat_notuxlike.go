//go:build !(unix || linux)

package cabrifsu

func HasFileWriteAccess(pathOrFi any) (bool, bool, error) {
	return true, true, nil
}
