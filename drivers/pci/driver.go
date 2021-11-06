package pci

var drivers = map[string]Driver{}

type Driver interface {
	Name() string
	Init(dev *Device) error
	Idents() []Identity
	Intr()
}

func Register(driver Driver) {
	drivers[driver.Name()] = driver
}
