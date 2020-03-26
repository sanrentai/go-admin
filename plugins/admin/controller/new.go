package controller

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/file"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	template2 "html/template"
	"net/http"
)

// ShowNewForm show a new form page.
func (h *Handler) ShowNewForm(ctx *context.Context) {
	param := guard.GetShowNewFormParam(ctx)
	h.showNewForm(ctx, "", param.Prefix, param.Param.GetRouteParamStr(), false)
}

func (h *Handler) showNewForm(ctx *context.Context, alert template2.HTML, prefix, paramStr string, isNew bool) {

	user := auth.Auth(ctx)

	panel := h.table(prefix, ctx)

	formInfo := panel.GetNewForm()

	infoUrl := h.routePathWithPrefix("info", prefix) + paramStr
	newUrl := h.routePathWithPrefix("new", prefix)
	showNewUrl := h.routePathWithPrefix("show_new", prefix) + paramStr

	referer := ctx.Headers("Referer")

	if referer != "" && !isInfoUrl(referer) && !isNewUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	hasAnimation := alert == ""
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: alert + formContent(aForm().
			SetPrefix(h.config.PrefixFixSlash()).
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetUrl(newUrl).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetHiddenFields(map[string]string{
				form2.TokenKey:    h.authSrv().AddToken(),
				form2.PreviousKey: infoUrl,
			}).
			SetTitle("New").
			SetOperationFooter(formFooter("new")).
			SetHeader(panel.GetForm().HeaderHtml).
			SetFooter(panel.GetForm().FooterHtml)),
		Description: panel.GetForm().Description,
		Title:       panel.GetForm().Title,
	}, h.config, menu.GetGlobalMenu(user, h.conn).SetActiveClass(h.config.URLRemovePrefix(ctx.Path())), hasAnimation)
	ctx.HTML(http.StatusOK, buf.String())

	if isNew {
		ctx.AddHeader(constant.PjaxUrlHeader, showNewUrl)
	}
}

// NewForm insert a table row into database.
func (h *Handler) NewForm(ctx *context.Context) {

	param := guard.GetNewFormParam(ctx)

	if param.HasAlert() {
		h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.GetRouteParamStr(), true)
		return
	}

	// process uploading files, only support local storage
	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			alert := aAlert().SetTitle(constant.DefaultErrorMsg).
				SetTheme("warning").
				SetContent(template2.HTML(err.Error())).
				GetContent()
			h.showNewForm(ctx, alert, param.Prefix, param.Param.GetRouteParamStr(), true)
			return
		}
	}

	err := param.Panel.InsertData(param.Value())
	if err != nil {
		alert := aAlert().SetTitle(constant.DefaultErrorMsg).
			SetTheme("warning").
			SetContent(template2.HTML(err.Error())).
			GetContent()
		h.showNewForm(ctx, alert, param.Prefix, param.Param.GetRouteParamStr(), true)
		return
	}

	if !param.FromList {

		if isNewUrl(param.PreviousPath, param.Prefix) {
			h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.GetRouteParamStr(), true)
			return
		}

		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	buf := h.showTable(ctx, param.Prefix, param.Param)

	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePathWithPrefix("info", param.Prefix)+param.Param.GetRouteParamStr())
}
