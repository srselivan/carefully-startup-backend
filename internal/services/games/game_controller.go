package games

type GameController struct {
	notify []func(bool)
}

func (ctrl *GameController) RegisterNotify(f func(bool)) {
	ctrl.notify = append(ctrl.notify, f)
}

func (ctrl *GameController) StartRegistrationPeriod() {
	for _, fn := range ctrl.notify {
		fn(true)
	}
}

func (ctrl *GameController) StopRegistrationPeriod() {
	for _, fn := range ctrl.notify {
		fn(false)
	}
}
