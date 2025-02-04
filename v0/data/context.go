package data

import (
	"context"
	"github.com/viant/datly/v0/db"
)

//Context visit context
type Context struct {
	context.Context
	View *View
	Db   db.Service
}

func NewContext(ctx context.Context, view *View, db db.Service) *Context {
	return &Context{
		Context: ctx,
		View:    view,
		Db:      db,
	}
}
