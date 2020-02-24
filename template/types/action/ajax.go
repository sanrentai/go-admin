package action

import (
	"encoding/json"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type AjaxAction struct {
	BtnId     string
	Url       string
	Method    string
	Data      AjaxData
	Alert     bool
	AlertData AlertData
	Handlers  []context.Handler
}

type AlertData struct {
	Title              string `json:"title"`
	Type               string `json:"type"`
	ShowCancelButton   bool   `json:"showCancelButton"`
	ConfirmButtonColor string `json:"confirmButtonColor"`
	ConfirmButtonText  string `json:"confirmButtonText"`
	CloseOnConfirm     bool   `json:"closeOnConfirm"`
	CancelButtonText   string `json:"cancelButtonText"`
}

func Ajax(url string, handler types.Handler) *AjaxAction {
	return &AjaxAction{
		Url:      url,
		Method:   "post",
		Data:     NewAjaxData(),
		Handlers: context.Handlers{handler.Wrap()},
	}
}

func (ajax *AjaxAction) WithAlert(data ...AlertData) *AjaxAction {
	ajax.Alert = true
	if len(data) > 0 {
		ajax.AlertData = data[0]
	} else {
		ajax.AlertData = AlertData{
			Title:              "",
			Type:               "warning",
			ShowCancelButton:   true,
			ConfirmButtonColor: "#DD6B55",
			ConfirmButtonText:  language.Get("yes"),
			CloseOnConfirm:     false,
			CancelButtonText:   language.Get("cancel"),
		}
	}
	return ajax
}

func (ajax *AjaxAction) SetData(data map[string]interface{}) *AjaxAction {
	ajax.Data = ajax.Data.Add(data)
	return ajax
}

func (ajax *AjaxAction) SetUrl(url string) *AjaxAction {
	ajax.Url = url
	return ajax
}

func (ajax *AjaxAction) SetMethod(method string) *AjaxAction {
	ajax.Method = method
	return ajax
}

func (ajax *AjaxAction) GetCallbacks() context.Node {
	return context.Node{
		Path:     ajax.Url,
		Method:   ajax.Method,
		Handlers: ajax.Handlers,
		Value:    map[string]interface{}{constant.ContextNodeNeedAuth: 1},
	}
}

func (ajax *AjaxAction) SetBtnId(btnId string)   { ajax.BtnId = btnId }
func (ajax *AjaxAction) BtnClass() template.HTML { return "" }

func (ajax *AjaxAction) Js() template.JS {

	ajaxStatement := `$.ajax({
                            method: '` + ajax.Method + `',
                            url: "` + ajax.Url + `",
                            data: data,
                            success: function (data) { 
                                if (data.code === 0) {
                                    swal(data.msg, '', 'success');
                                } else {
                                    swal(data.msg, '', 'error');
                                }
                            }
                        });`

	if ajax.Alert {
		b, _ := json.Marshal(ajax.AlertData)
		ajaxStatement = "swal(" + string(b) + `,
                    function () {
						` + ajaxStatement + `
					});`
	}

	return template.JS(`$('.` + ajax.BtnId + `').on('click', function (event) {
						let data = ` + ajax.Data.JSON() + `;
						let id = $(this).attr("data-id");
						if (id && id !== "") {
							data["id"] = id;
						}
						` + ajaxStatement + `
            		});`)
}

func (ajax *AjaxAction) BtnAttribute() template.HTML { return template.HTML(`href="javascript:;"`) }
func (ajax *AjaxAction) ExtContent() template.HTML   { return template.HTML(``) }
