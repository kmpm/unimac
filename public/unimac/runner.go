package unimac

type runner struct {
	err error
}

func (r *runner) run(f func() error) {
	if r.err != nil {
		return
	}
	r.err = f()
}
