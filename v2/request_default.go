package command

import "time"

type DefaultRequest struct {
	Vars map[string]string
	Data []byte
}

func (r *DefaultRequest) GetVars() map[string]string {
	return r.Vars
}

func (r *DefaultRequest) GetData() []byte {
	return r.Data
}

func (r *DefaultRequest) SetDate(date time.Time) {
}
