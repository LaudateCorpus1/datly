package data

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs"
	cache2 "github.com/viant/datly/v0/cache"
	"github.com/viant/gtly"
	"github.com/viant/toolbox/data"
	"strings"
)

//View represents a data view
type View struct {
	Connector  string
	Name       string
	Alias      string        `json:",omitempty"`
	Table    string        `json:",omitempty"`
	From     *From         `json:",omitempty"`
	Criteria *Criteria     `json:",omitempty"`
	Selector Selector      `json:",omitempty"`
	Joins    []*Join       `json:",omitempty"`
	Refs     []*Reference  `json:",omitempty"`
	Params   []interface{} `json:",omitempty"`
	CaseFormat string        `json:",omitempty"`
	HideRefIDs bool          `json:",omitempty"`

	PrimaryKey    []string     `json:",omitempty"`
	Mutable       *bool        `json:",omitempty"`
	Columns       []*Column    `json:",omitempty"`
	Parameters    []*Parameter `json:",omitempty"`
	Cache         *Cache       `json:",omitempty"`
	OnRead        *Visitor     `json:",omitempty"`
	OnPath        *Visitor     `json:",omitempty"`
	_cacheService cache2.Service
}

//Cacher returns a cache service
func (v *View) Cacher() cache2.Service {
	return v._cacheService
}

//Clone creates a view clone
func (v *View) Clone() *View {
	return &View{
		Connector:  v.Connector,
		Name:       v.Name,
		Alias:      v.Alias,
		Table:      v.Table,
		PrimaryKey: v.PrimaryKey,
		Mutable:    v.Mutable,
		From:       v.From,
		Columns:    v.Columns,
		Parameters: v.Parameters,
		Criteria:   v.Criteria,
		Selector:   v.Selector,
		Refs:       v.Refs,
		Params:     v.Params,
		CaseFormat: v.CaseFormat,
		HideRefIDs: v.HideRefIDs,
		OnRead:     v.OnRead,
	}
}

//MergeFrom merges from template view
func (v *View) MergeFrom(tmpl *View) {
	if v.From == nil {
		v.From = tmpl.From
	}
	if v.Table == "" {
		v.Table = tmpl.Table
	}
	if len(v.PrimaryKey) == 0 {
		v.PrimaryKey = tmpl.PrimaryKey
	}
	if v.Mutable == nil {
		v.Mutable = tmpl.Mutable
	}
	if v.CaseFormat == "" {
		v.CaseFormat = tmpl.CaseFormat
	}
	if v.HideRefIDs {
		v.HideRefIDs = tmpl.HideRefIDs
	}
	if v.Alias == "" {
		v.Alias = tmpl.Alias
	}
	if v.Connector == "" {
		v.Connector = tmpl.Connector
	}

	if len(v.Columns) == 0 {
		v.Columns = tmpl.Columns
	}
	if len(v.Refs) == 0 {
		v.Refs = tmpl.Refs
	}
	if len(v.Parameters) == 0 {
		v.Parameters = tmpl.Parameters
	}
	if len(v.Joins) == 0 {
		v.Joins = tmpl.Joins
	}
	if len(v.Params) == 0 {
		v.Params = tmpl.Params
	}
	if v.Criteria == nil {
		v.Criteria = tmpl.Criteria
	}
	if v.OnRead == nil {
		v.OnRead = tmpl.OnRead
	}
}

//IsMutable returns true if mutable
func (v *View) IsMutable() bool {
	if v.Mutable == nil {
		return false
	}
	return *v.Mutable
}

//AddJoin add join
func (v *View) AddJoin(join *Join) {
	v.Joins = append(v.Joins, join)
}

//LoadSQL loads fromSQL
func (v *View) LoadSQL(ctx context.Context, fs afs.Service, parentURL string) error {
	if v.From == nil {
		return nil
	}
	if err := v.From.Fragment.LoadSQL(ctx, fs, parentURL); err != nil {
		return err
	}

	var fragments = data.NewMap()
	if len(v.From.Fragments) > 0 {
		for _, fragment := range v.From.Fragments {
			if err := fragment.LoadSQL(ctx, fs, parentURL); err != nil {
				return err
			}
			fragments[fragment.Key] = fragments.ExpandAsText(fragment.SQL)
		}
	}
	v.From.SQL = fragments.ExpandAsText(v.From.SQL)
	return nil
}

//Validate checks if view is valid
func (v View) Validate() error {
	if v.Table == "" && (v.From == nil || v.From.SQL == "") {
		return errors.Errorf("table was empty")
	}
	if v.Connector == "" {
		return errors.Errorf("connector was empty")
	}
	if len(v.Parameters) > 0 {
		for i := range v.Parameters {
			if err := v.Parameters[i].Validate(); err != nil {
				return err
			}
		}
	}
	if len(v.Refs) > 0 {
		for i := range v.Refs {
			if err := v.Refs[i].Validate(); err != nil {
				return err
			}
		}
	}
	if v.CaseFormat != "" {
		if err := gtly.ValidateCaseFormat(v.CaseFormat); err != nil {
			return errors.Wrapf(err, "invalid view: %v", v.Name)
		}
	}

	if v.Mutable != nil && *v.Mutable {
		if len(v.PrimaryKey) == 0 {
			return errors.Errorf("primaryKey was empty on data view %v", v.Name)
		}
		if v.Table == "" {
			return errors.Errorf("table was empty on data view %v", v.Name)
		}
	}

	return nil
}

//Init initializes view
func (v *View) Init(setPrefix bool) error {

	if v.Name == "" && v.Table != "" {
		v.Name = v.Table
	}
	if v.Alias == "" {
		v.Alias = "t"
	}
	if setPrefix && v.Selector.Prefix == "" {
		v.Selector.Prefix = v.Name
	}
	if len(v.Parameters) > 0 {
		for i := range v.Parameters {
			v.Parameters[i].Init()
		}
	}

	if v.OnRead != nil {
		if err := v.OnRead.Init(); err != nil {
			return err
		}
	}
	if v.Cache != nil && v.Cache.Service != "" {
		var err error
		if v._cacheService, err = cache2.Registry().Get(v.Cache.Service); err != nil {
			return err
		}
		v.Cache.Init()
	}
	//If primary key is specified set mutable flag be default
	if isMutable := len(v.PrimaryKey) > 0; isMutable && v.Mutable == nil {
		v.Mutable = &isMutable
	}
	return nil
}

const (
	projectionKey = "projection"
	fromKey       = "from"
	aliasKey      = "alias"
	criteriaKey   = "criteria"
	joinsKey      = "joins"
	limitKey      = "limit"
	orderByKey    = "orderBy"
	sqlTemplate   = `SELECT $projection 
FROM $from ${alias}${joins}${criteria}${orderBy}${limit}`
)

//BuildSQL build data view FromFragments
func (v View) BuildSQL(selector *Selector, bindings map[string]interface{}) (string, []interface{}, error) {
	projection := v.buildSQLProjection(selector)
	from := v.buildSQLFom(bindings)
	orderBy := v.buildSQLOrderBy(selector)
	criteria, parameters := v.buildSQLCriteria(selector, bindings)
	limit := v.buildSQLLimit(selector, bindings)
	joins := v.buildSQLJoins(selector, bindings)

	var replacements = data.NewMap()
	replacements.Put(projectionKey, projection)
	replacements.Put(fromKey, from)
	replacements.Put(aliasKey, v.Alias)
	replacements.Put(criteriaKey, criteria)
	replacements.Put(limitKey, limit)
	replacements.Put(orderByKey, orderBy)
	replacements.Put(joinsKey, joins)
	SQL := replacements.ExpandAsText(sqlTemplate)
	return SQL, parameters, nil
}

func (v View) buildSQLFom(bindings data.Map) string {
	from := v.Table
	if v.From != nil && v.From.SQL != "" {
		from = "(" + v.From.SQL + ")"
	}
	return bindings.ExpandAsText(from)
}

func (v View) buildSQLProjection(selector *Selector) string {
	projection := v.Alias + ".*"

	if len(selector.Columns) > 0 {
		var columns = make([]string, 0)
		for i := range selector.Columns {
			columns = append(columns, fmt.Sprintf("\t%v.%v", v.Alias, selector.Columns[i]))
		}
		projection = strings.Join(columns, ",\n")
	}
	return projection
}

func (v View) buildSQLOrderBy(selector *Selector) string {
	if selector.OrderBy == "" {
		return ""
	}
	return "\nORDER BY " + selector.OrderBy
}

func (v View) buildSQLCriteria(selector *Selector, bindings data.Map) (string, []interface{}) {
	result := ""
	if v.Criteria == nil && selector.Criteria == nil {
		return result, nil
	}
	var parameters = make([]interface{}, 0)
	if v.Criteria != nil {
		result = "\nWHERE " + expendCriteria(bindings, v.Criteria)
		parameters = addCriteriaParams(v.Criteria.Params, bindings, parameters)
	}

	if selector.Criteria == nil {
		return result, parameters
	}
	if result == "" {
		result += "\nWHERE "
	} else {
		result += "\n AND "
	}
	result += expendCriteria(bindings, selector.Criteria)
	parameters = addCriteriaParams(selector.Criteria.Params, bindings, parameters)
	return result, parameters
}

func expendCriteria(bindings data.Map, criteria *Criteria) string {
	expression := bindings.ExpandAsText(criteria.Expression)
	if !strings.Contains(expression, "=") {
		expression = strings.Replace(expression, ":", "=", len(expression))
	}
	return "(" + bindings.ExpandAsText(expression) + ")"
}

func addCriteriaParams(nameParams []string, bindings data.Map, valueParams []interface{}) []interface{} {
	if len(nameParams) == 0 {
		return valueParams
	}
	for _, key := range nameParams {
		value, ok := bindings[key]
		if !ok {
			value, _ = bindings.GetValue(key)
		}
		valueParams = append(valueParams, value)
	}
	return valueParams
}

func (v View) buildSQLLimit(selector *Selector, bindings map[string]interface{}) string {
	if selector.Limit == 0 && selector.Offset == 0 {
		return ""
	}
	result := ""
	if selector.Limit > 0 {
		result = fmt.Sprint(" LIMIT  ", selector.Limit)
	}
	if selector.Offset > 0 {
		result += fmt.Sprint(" OFFSET  ", selector.Offset)
	}
	return result
}

func (v *View) buildSQLJoins(selector *Selector, bindings map[string]interface{}) string {
	if len(v.Joins) == 0 {
		return ""
	}
	var joins = make([]string, 0)
	for _, join := range v.Joins {
		joins = append(joins, fmt.Sprintf(" %v JOIN %v %v ON %v", join.Type, join.Table, join.Alias, join.On))
	}
	return "\n" + strings.Join(joins, "\n")
}
