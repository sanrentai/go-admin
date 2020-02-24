package table

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types"
	"net/http"
	"net/url"
)

type Generator func(ctx *context.Context) Table

type GeneratorList map[string]Generator

func (g GeneratorList) InjectRoutes(app *context.App, srv service.List) {
	authHandler := auth.Middleware(db.GetConnection(srv))
	for _, gen := range g {
		table := gen(context.NewContext(&http.Request{
			URL: &url.URL{},
		}))
		for _, cb := range table.GetInfo().Callbacks {
			if cb.Value[constant.ContextNodeNeedAuth] == 1 {
				app.AppendReqAndResp(cb.Path, cb.Method, append([]context.Handler{authHandler}, cb.Handlers...))
			} else {
				app.AppendReqAndResp(cb.Path, cb.Method, cb.Handlers)
			}
		}
		for _, cb := range table.GetForm().Callbacks {
			if cb.Value[constant.ContextNodeNeedAuth] == 1 {
				app.AppendReqAndResp(cb.Path, cb.Method, append([]context.Handler{authHandler}, cb.Handlers...))
			} else {
				app.AppendReqAndResp(cb.Path, cb.Method, cb.Handlers)
			}
		}
	}
}

func (g GeneratorList) Add(key string, gen Generator) {
	g[key] = gen
}

func (g GeneratorList) Combine(gg GeneratorList) {
	for key, gen := range gg {
		if _, ok := g[key]; !ok {
			g[key] = gen
		}
	}
}

func (g GeneratorList) CombineAll(ggg []GeneratorList) {
	for _, gg := range ggg {
		for key, gen := range gg {
			if _, ok := g[key]; !ok {
				g[key] = gen
			}
		}
	}
}

var generators = make(GeneratorList)

func Get(key string, ctx *context.Context) Table {
	return generators[key](ctx)
}

// SetGenerators update generators.
func SetGenerators(gens map[string]Generator) {
	for key, gen := range gens {
		generators[key] = gen
	}
}

type Table interface {
	GetInfo() *types.InfoPanel
	GetDetail() *types.InfoPanel
	GetForm() *types.FormPanel

	GetCanAdd() bool
	GetEditable() bool
	GetDeletable() bool
	GetExportable() bool
	IsShowDetail() bool

	GetPrimaryKey() PrimaryKey

	GetData(path string, params parameter.Parameters, isAll bool) (PanelInfo, error)
	GetDataWithIds(path string, params parameter.Parameters, ids []string) (PanelInfo, error)
	GetDataWithId(id string) ([]types.FormField, [][]types.FormField, []string, string, string, error)
	UpdateDataFromDatabase(dataList form.Values) error
	InsertDataFromDatabase(dataList form.Values) error
	DeleteDataFromDatabase(id string) error

	Copy() Table
}

type PrimaryKey struct {
	Type db.DatabaseType
	Name string
}

const (
	DefaultPrimaryKeyName = "id"
	DefaultConnectionName = "default"
)

var services service.List

func SetServices(srv service.List) {
	services = srv
}

// sql is a helper function return db sql.
func (tb DefaultTable) sql() *db.SQL {
	return db.WithDriverAndConnection(tb.connection, db.GetConnectionFromService(services.Get(tb.connectionDriver)))
}

func GetNewFormList(groupHeaders []string,
	group [][]string,
	old []types.FormField) ([]types.FormField, [][]types.FormField, []string) {

	if len(group) == 0 {
		var newForm []types.FormField
		for _, v := range old {
			v.Value = v.Default
			if !v.NotAllowAdd {
				v.Editable = true
				newForm = append(newForm, v)
			}
		}
		return newForm, [][]types.FormField{}, []string{}
	}

	var (
		newForm = make([][]types.FormField, 0)
		headers = make([]string, 0)
	)

	for key, value := range group {
		list := make([]types.FormField, 0)

		for i := 0; i < len(value); i++ {
			for _, v := range old {
				if v.Field == value[i] {
					v.Value = v.Default
					if !v.NotAllowAdd {
						v.Editable = true
						list = append(list, v)
						break
					}
				}
			}
		}

		newForm = append(newForm, list)
		headers = append(headers, groupHeaders[key])
	}

	return []types.FormField{}, newForm, headers
}

// ***************************************
// helper function for database operation
// ***************************************

func delimiter(del, s string) string {
	if del == "[" {
		return "[" + s + "]"
	}
	return del + s + del
}

func filterFiled(filed, delimiter string) string {
	if delimiter == "[" {
		return filed
	}
	return delimiter + filed + delimiter
}

type Columns []string

func (tb DefaultTable) getColumns(columnsModel []map[string]interface{}) (Columns, bool) {
	columns := make(Columns, len(columnsModel))
	switch tb.connectionDriver {
	case "postgresql":
		auto := false
		for key, model := range columnsModel {
			columns[key] = model["column_name"].(string)
			if columns[key] == tb.primaryKey.Name {
				if v, ok := model["column_default"].(string); ok {
					if strings.Contains(v, "nextval") {
						auto = true
					}
				}
			}
		}
		return columns, auto
	case "mysql":
		auto := false
		for key, model := range columnsModel {
			columns[key] = model["Field"].(string)
			if columns[key] == tb.primaryKey.Name {
				if v, ok := model["Extra"].(string); ok {
					if v == "auto_increment" {
						auto = true
					}
				}
			}
		}
		return columns, auto
	case "sqlite":
		for key, model := range columnsModel {
			columns[key] = string(model["name"].(string))
		}

		num, _ := tb.sql().Table("sqlite_sequence").
			Where("name", "=", tb.GetForm().Table).Count()

		return columns, num > 0
	case "mssql":
		for key, model := range columnsModel {
			columns[key] = string(model["column_name"].(string))
		}
		return columns, true
	default:
		panic("wrong driver")
	}
}

func getAggregationExpression(driver, field, headField, delimiter string) string {
	switch driver {
	case "postgresql":
		return fmt.Sprintf("string_agg(%s::character varying, '%s') as %s", field, delimiter, headField)
	case "mysql":
		return fmt.Sprintf("group_concat(%s separator '%s') as %s", field, delimiter, headField)
	case "sqlite":
		return fmt.Sprintf("group_concat(%s, '%s') as %s", field, delimiter, headField)
	default:
		panic("wrong driver")
	}
}

// inArray checks the find string is in the columns or not.
func inArray(columns []string, find string) bool {
	for i := 0; i < len(columns); i++ {
		if columns[i] == find {
			return true
		}
	}
	return false
}
