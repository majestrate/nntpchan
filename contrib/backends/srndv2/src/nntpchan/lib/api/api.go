package api

import (
	"nntpchan/lib/model"
)
// json api
type API interface {
	MakePost(p model.Post)
}
