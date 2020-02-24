package form

import (
	"github.com/GoAdminGroup/go-admin/modules/db"
	"html/template"
)

type Type uint8

const (
	Default Type = iota
	Text
	SelectSingle
	Select
	IconPicker
	SelectBox
	File
	Password
	RichText
	Datetime
	DatetimeRange
	Radio
	Email
	Url
	Ip
	Color
	Currency
	Number
	NumberRange
	TextArea
	Custom
	Switch
)

type Layout uint8

const (
	LayoutDefault Layout = iota
	LayoutTwoCol
	LayoutFlow
	LayoutTab
)

func (t Type) String() string {
	switch t {
	case Default:
		return "default"
	case Text:
		return "text"
	case SelectSingle:
		return "select_single"
	case Select:
		return "select"
	case IconPicker:
		return "iconpicker"
	case SelectBox:
		return "selectbox"
	case File:
		return "file"
	case Password:
		return "password"
	case RichText:
		return "richtext"
	case Datetime:
		return "datetime"
	case DatetimeRange:
		return "datetime_range"
	case Radio:
		return "radio"
	case Email:
		return "email"
	case Url:
		return "url"
	case Ip:
		return "ip"
	case Color:
		return "color"
	case Currency:
		return "currency"
	case Number:
		return "number"
	case NumberRange:
		return "number_range"
	case TextArea:
		return "textarea"
	case Custom:
		return "custom"
	case Switch:
		return "switch"
	default:
		panic("wrong form type")
	}
}

func (t Type) IsSelect() bool {
	return t == Select || t == SelectSingle || t == SelectBox || t == Radio || t == Switch
}

func (t Type) IsMultiSelect() bool {
	return t == Select || t == SelectBox
}

func (t Type) IsRange() bool {
	return t == DatetimeRange || t == NumberRange
}

func (t Type) SelectedLabel() []string {
	if t == Select || t == SelectSingle || t == SelectBox {
		return []string{"selected", ""}
	}
	if t == Radio || t == Switch {
		return []string{"checked", ""}
	}
	return []string{}
}

func GetFormTypeFromFieldType(typeName db.DatabaseType, fieldName string) string {

	if fieldName == "password" {
		return "form.Password"
	}

	if fieldName == "id" {
		return "form.Default"
	}

	if fieldName == "ip" {
		return "form.Ip"
	}

	if fieldName == "Url" {
		return "form.Url"
	}

	if fieldName == "email" {
		return "form.Email"
	}

	if fieldName == "color" {
		return "form.Color"
	}

	if fieldName == "money" {
		return "form.Currency"
	}

	switch typeName {
	case db.Int, db.Tinyint, db.Int4, db.Integer, db.Mediumint, db.Smallint,
		db.Numeric, db.Smallserial, db.Serial, db.Bigserial, db.Money, db.Bigint:
		return "form.Number"
	case db.Text, db.Longtext, db.Mediumtext, db.Tinytext:
		return "form.RichText"
	case db.Datetime, db.Date, db.Time, db.Timestamp, db.Timestamptz, db.Year:
		return "form.Datetime"
	}

	return "form.Text"
}

func DefaultHTML(value string) template.HTML {
	return template.HTML(`<div class="box box-solid box-default no-margin"><div class="box-body" style="min-height: 40px;">` +
		value + `</div></div>`)
}

func HiddenInputHTML(field, value string) template.HTML {
	return template.HTML(`<input type="hidden" name="` + field + `" value="` + value + `" class="form-control">`)
}
