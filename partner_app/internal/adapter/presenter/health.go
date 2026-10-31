package presenter

type HealthPresenter struct{}

func NewHealthPresenter(opts ...Option) *HealthPresenter { return &HealthPresenter{} }

type healthBody struct {
	Status string `json:"status"`
}

func (p *HealthPresenter) CheckLiveness() healthBody { return healthBody{Status: "ok"} }

func (p *HealthPresenter) CheckReadiness() healthBody { return healthBody{Status: "ready"} }
