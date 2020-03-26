package controller

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	template2 "html/template"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
)

// GlobalDeferHandler is a global error handler of admin plugin.
func (h *Handler) GlobalDeferHandler(ctx *context.Context) {

	logger.Access(ctx)

	h.RecordOperationLog(ctx)

	if err := recover(); err != nil {
		logger.Error(err)
		logger.Error(string(debug.Stack()[:]))

		var (
			errMsg string
			ok     bool
			e      error
		)

		if errMsg, ok = err.(string); !ok {
			if e, ok = err.(error); ok {
				errMsg = e.Error()
			}
		}

		if errMsg == "" {
			errMsg = "system error"
		}

		if ok, _ = regexp.MatchString("/edit(.*)", ctx.Path()); ok {
			h.setFormWithReturnErrMessage(ctx, errMsg, "edit")
			return
		}
		if ok, _ = regexp.MatchString("/new(.*)", ctx.Path()); ok {
			h.setFormWithReturnErrMessage(ctx, errMsg, "new")
			return
		}

		alert := aAlert().
			SetTitle(constant.DefaultErrorMsg).
			SetTheme("warning").
			SetContent(template2.HTML(errMsg)).
			GetContent()

		user := auth.Auth(ctx)

		tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
		buf := template.Execute(tmpl, tmplName, user, types.Panel{
			Content:     alert,
			Description: "error",
			Title:       "error",
		}, h.config, menu.GetGlobalMenu(user, h.conn).SetActiveClass(h.config.URLRemovePrefix(ctx.Path())))
		ctx.HTML(http.StatusOK, buf.String())
		return
	}
}

func (h *Handler) setFormWithReturnErrMessage(ctx *context.Context, errMsg string, kind string) {

	alert := aAlert().
		SetTitle(constant.DefaultErrorMsg).
		SetTheme("warning").
		SetContent(template2.HTML(errMsg)).
		GetContent()

	var (
		formInfo table.FormInfo
		prefix   = ctx.Query(constant.PrefixKey)
		panel    = h.table(prefix, ctx)
	)

	if kind == "edit" {
		id := ctx.Query("id")
		if id == "" {
			id = ctx.Request.MultipartForm.Value[panel.GetPrimaryKey().Name][0]
		}
		formInfo, _ = h.table(prefix, ctx).GetDataWithId(parameter.GetParam(ctx.Request.URL,
			panel.GetInfo().DefaultPageSize,
			panel.GetInfo().SortField,
			panel.GetInfo().GetSort()).WithPKs(id))
	} else {
		formInfo = panel.GetNewForm()
		formInfo.Title = panel.GetForm().Title
		formInfo.Description = panel.GetForm().Description
	}

	queryParam := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize,
		panel.GetInfo().SortField, panel.GetInfo().GetSort()).GetRouteParamStr()

	user := auth.Auth(ctx)

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: alert + formContent(aForm().
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetTitle(template2.HTML(strings.Title(kind))).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetPrefix(h.config.PrefixFixSlash()).
			SetHiddenFields(map[string]string{
				form.TokenKey:    h.authSrv().AddToken(),
				form.PreviousKey: h.config.Url("/info/" + prefix + queryParam),
			}).
			SetUrl(h.config.Url("/"+kind+"/"+prefix)).
			SetOperationFooter(formFooter(kind)).
			SetHeader(panel.GetForm().HeaderHtml).
			SetFooter(panel.GetForm().FooterHtml)),
		Description: formInfo.Description,
		Title:       formInfo.Title,
	}, h.config, menu.GetGlobalMenu(user, h.conn).SetActiveClass(h.config.URLRemovePrefix(ctx.Path())))
	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, h.config.Url("/info/"+prefix+"/"+kind+queryParam))
}
