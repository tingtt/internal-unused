package consumer

import "example.com/multimode/internal/api"

func Run() {
	api.ProdUsed()
	_ = api.UsedType{Active: true}
	_ = api.UnusedMemberType{}
}
