package tools

type GenericListContent struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

func (l GenericListContent) GetName() string {
	return l.Name
}

func (l GenericListContent) GetNamespace() string {
	return l.Namespace
}
