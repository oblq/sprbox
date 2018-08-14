package global

type parsable interface {
	SpareConfig([]string) error
}
