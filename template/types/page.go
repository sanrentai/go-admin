// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/modules/system"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"html/template"
	"strconv"
)

// Attribute is the component interface of template. Every component of
// template should implement it.
type Attribute struct {
	TemplateList map[string]string
}

// Page used in the template as a top variable.
type Page struct {
	// User is the login user.
	User models.UserModel

	// Menu is the left side menu of the template.
	Menu menu.Menu

	// Panel is the main content of template.
	Panel Panel

	// System contains some system info.
	System SystemInfo

	// UrlPrefix is the prefix of url.
	UrlPrefix string

	// Title is the title of the web page.
	Title string

	// Logo is the logo of the template.
	Logo template.HTML

	// MiniLogo is the downsizing logo of the template.
	MiniLogo template.HTML

	// ColorScheme is the color scheme of the template.
	ColorScheme string

	// IndexUrl is the home page url of the site.
	IndexUrl string

	// AssetUrl is the cdn link of assets
	CdnUrl string

	// Custom html in the tag head.
	CustomHeadHtml template.HTML

	// Custom html after body.
	CustomFootHtml template.HTML

	// Components assets
	AssetsList template.HTML
}

func NewPage(user models.UserModel, menu menu.Menu, panel Panel, cfg config.Config, assetsList template.HTML) Page {
	return Page{
		User:  user,
		Menu:  menu,
		Panel: panel,
		System: SystemInfo{
			Version: system.Version(),
		},
		UrlPrefix:      cfg.AssertPrefix(),
		Title:          cfg.Title,
		Logo:           cfg.Logo,
		MiniLogo:       cfg.MiniLogo,
		ColorScheme:    cfg.ColorScheme,
		IndexUrl:       cfg.GetIndexURL(),
		CdnUrl:         cfg.AssetUrl,
		CustomHeadHtml: cfg.CustomHeadHtml,
		CustomFootHtml: cfg.CustomFootHtml,
		AssetsList:     assetsList,
	}
}

// SystemInfo contains basic info of system.
type SystemInfo struct {
	Version string
}

// Panel contains the main content of the template which used as pjax.
type Panel struct {
	Title       string
	Description string
	Content     template.HTML
	Url         string

	// Whether to toggle the sidebar
	MiniSidebar bool

	// Auto refresh page switch.
	AutoRefresh bool
	// Refresh page intervals, the unit is second.
	RefreshInterval []int
}

func (p Panel) GetContent(prod bool) Panel {
	if p.MiniSidebar {
		p.Content += `<script>$("body").addClass("sidebar-collapse")</script>`
	}
	if p.AutoRefresh {
		refreshTime := 60
		if len(p.RefreshInterval) > 0 {
			refreshTime = p.RefreshInterval[0]
		}

		p.Content += `<script>
window.setTimeout(function(){
	$.pjax.reload('#pjax-container');
}, ` + template.HTML(strconv.Itoa(refreshTime*1000)) + `);
</script>`
	}
	if prod {
		utils.CompressedContent(&p.Content)
	}
	return p
}

type GetPanelFn func(ctx interface{}) (Panel, error)
