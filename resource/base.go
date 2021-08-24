package resource

type Store interface {
	Save(r Resource) error
	Load(r Resource) error
}

type Resource interface {
	Hash() []byte
	GetData() []byte
	SetData([]byte)
	IsStored() bool
	SetStored()
	GetLocation() string
	SetLocation(l string)
	GetID() string
}

type BaseResource struct {
	Resource
	Stored   bool
	Location string
	ID       string
}

func (r *BaseResource) IsStored() bool {
	return r.Stored
}

func (r *BaseResource) SetStored() {
	r.Stored = true
}

func (r *BaseResource) GetLocation() string {
	return r.Location
}

func (r *BaseResource) SetLocation(l string) {
	r.Location = l
}

func (r *BaseResource) GetID() string {
	return r.ID
}

func (r *BaseResource) SetID(id string) {
	r.ID = id
}
