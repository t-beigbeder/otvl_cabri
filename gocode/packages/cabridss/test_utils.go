package cabridss

func GetServerIndexesForTests(dss HDss) (Index, Index) {
	if dss == nil {
		return nil, nil
	}
	px, ok := dss.(*ODss).proxy.(*webDssImpl)
	if !ok {
		px, ok := dss.(*ODss).proxy.(*eDssImpl)
		if !ok {
			return nil, nil
		}
		if !px.libApi {
			return nil, nil
		}
		return px.apc.GetConfig().(webDssClientConfig).libDss.GetIndex(), dss.GetIndex()
	}
	if !px.libApi {
		return nil, nil
	}
	return px.apc.GetConfig().(webDssClientConfig).libDss.GetIndex(), dss.GetIndex()
}
