package helper

func ArrayToMap(s []string) map[string]struct{} {
	res := map[string]struct{}{}
	for _, v := range s {
		res[v] = struct{}{}
	}

	return res
}
