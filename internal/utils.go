package internal

func FilterNils(errs []error) []error {
	var r []error
	for _, v := range errs {
		if v != nil {
			r = append(r, v)
		}
	}
	return r
}
