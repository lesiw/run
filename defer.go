package main

type deferlist []func()

func (d *deferlist) add(f func()) {
	*d = append(deferlist{f}, *d...)
}

func (d *deferlist) run() {
	for _, f := range *d {
		f()
	}
}
