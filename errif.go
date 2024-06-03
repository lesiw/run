package main

func errIf(stmt bool, fn func() error) error {
	if stmt {
		return fn()
	} else {
		return nil
	}
}
