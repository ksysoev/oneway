package connapi

type API struct{}

type Config struct {
	Listen string
}

func New(cfg Config) *API {
	return &API{}
}

func (a *API) Run() error {
	return nil
}
