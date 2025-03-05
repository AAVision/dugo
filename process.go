package main

func groupByHash(files []string) ([]string, error) {
	m := map[string][]string{}

	for _, v := range files {
		h, err := createFileHash(v)
		if err != nil {
			return nil, err
		}
		m[h] = append(m[h], v)
	}

	s := []string{}

	for _, v := range m {
		if len(v) > 1 {
			s = append(s, v...)
		}
	}

	return s, nil
}
