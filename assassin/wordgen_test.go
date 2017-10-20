package assassin

type dummyWordGen struct {
	b byte
}

func (d *dummyWordGen) Gen() string {
	var c = string(d.b)
	d.b++
	return c + c + c + c
}
