package resource

type Cache interface {
	Store(r Resource) error
	Load(r Resource) error
}

type Resource struct {
}
