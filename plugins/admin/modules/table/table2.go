package table

import (
	"github.com/GoAdminGroup/go-admin/context"
)

func RefreshTable(key string,ctx *context.Context) {
	tableList[key] = generators[key](ctx)
}