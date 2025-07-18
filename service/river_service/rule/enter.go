// Path: ./service/river_service/rule/enter.go

package rule

import (
	"github.com/siddontang/go-mysql/schema"
	"strings"
)

// Rule is the rule for how to sync data from MySQL to ES.
// If you want to sync MySQL data into elasticsearch, you must set a rule to let use know how to do it.
// The mapping rule may thi: schema + table <-> index + document type.
// schema and table is for MySQL, index and document type is for Elasticsearch.
type Rule struct {
	Schema string   `yaml:"schema"`
	Table  string   `yaml:"table"`
	Index  string   `yaml:"index"`
	Type   string   `yaml:"type"`
	Parent string   `yaml:"-"`
	ID     []string `yaml:"-"`

	// Default, a MySQL table field name is mapped to Elasticsearch field name.
	// Sometimes, you want to use different name, e.g, the MySQL file name is title,
	// but in Elasticsearch, you want to name it my_title.
	FieldMapping map[string]string `yaml:"field"`

	// MySQL table information
	TableInfo *schema.Table `yaml:"-"`

	//only MySQL fields in filter will be synced , default sync all fields
	Filter []string `yaml:"-"`

	// Elasticsearch pipeline
	// To pre-process documents before indexing
	Pipeline string `yaml:"-"`
}

func NewDefaultRule(schema string, table string) *Rule {
	r := new(Rule)

	r.Schema = schema
	r.Table = table

	lowerTable := strings.ToLower(table)
	r.Index = lowerTable
	r.Type = lowerTable

	r.FieldMapping = make(map[string]string)

	return r
}

func (r *Rule) Prepare() error {
	if r.FieldMapping == nil {
		r.FieldMapping = make(map[string]string)
	}

	if len(r.Index) == 0 {
		r.Index = r.Table
	}

	if len(r.Type) == 0 {
		r.Type = r.Index
	}

	// ES must use a lower-case Type
	// Here we also use for Index
	r.Index = strings.ToLower(r.Index)
	r.Type = strings.ToLower(r.Type)

	return nil
}

// CheckFilter checkers whether the field needs to be filtered.
func (r *Rule) CheckFilter(field string) bool {
	if r.Filter == nil {
		return true
	}

	for _, f := range r.Filter {
		if f == field {
			return true
		}
	}
	return false
	//_, exists := r.FieldMapping[field]
	//return exists

}
