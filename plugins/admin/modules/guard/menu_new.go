package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"html/template"
	"strconv"
)

type MenuNewParam struct {
	Title    string
	Header   string
	ParentId int64
	Icon     string
	Uri      string
	Roles    []string
	Alert    template.HTML
}

func (e MenuNewParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

func (g *Guard) MenuNew(ctx *context.Context) {

	parentId := ctx.FormValue("parent_id")
	if parentId == "" {
		parentId = "0"
	}

	var (
		alert template.HTML
		token = ctx.FormValue(form.TokenKey)
	)

	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert = getAlert("edit fail, wrong token")
	}

	if alert == "" {
		alert = checkEmpty(ctx, "title", "icon")
	}

	parentIdInt, _ := strconv.Atoi(parentId)

	ctx.SetUserValue("new_menu_param", &MenuNewParam{
		Title:    ctx.FormValue("title"),
		Header:   ctx.FormValue("header"),
		ParentId: int64(parentIdInt),
		Icon:     ctx.FormValue("icon"),
		Uri:      ctx.FormValue("uri"),
		Roles:    ctx.Request.Form["roles[]"],
		Alert:    alert,
	})
	ctx.Next()
}

func GetMenuNewParam(ctx *context.Context) *MenuNewParam {
	return ctx.UserValue["new_menu_param"].(*MenuNewParam)
}
