package zkWrap

func pathWithChroot(path string) string {
	return root + path
}

func stringSlice2InterfaceSlice(before []string) (after []interface{}) {
	for _, v := range before {
		after = append(after, v)
	}
	return
}

func interfaceSlice2StringSlice(before []interface{}) (after []string) {
	for _, v := range before {
		after = append(after, v.(string))
	}
	return
}
