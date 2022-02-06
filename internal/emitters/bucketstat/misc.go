package bucketstat

type objSizeInfo struct {
	size    int64
	deleted bool
}

type objSizeMap map[string]*objSizeInfo

func (s objSizeMap) get(key string) *objSizeInfo {
	obj, ok := s[key]
	if !ok {
		obj = &objSizeInfo{}
		s[key] = obj
	}
	return obj
}
