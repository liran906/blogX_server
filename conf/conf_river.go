// Path: ./conf/conf_river.go

package conf

import (
	"blogX_server/service/river_service/rule"
)

type River struct {
	Enable   bool                `yaml:"enable"`
	ServerID uint32              `yaml:"serverID"`
	Flavor   string              `yaml:"flavor"`
	DataDir  string              `yaml:"dataDir"`
	Sources  []RiverSourceConfig `yaml:"sources"`
	Rules    []*rule.Rule        `yaml:"rule"`
	BulkSize int                 `yaml:"bulkSize"`
}

type RiverSourceConfig struct {
	Schema string   `yaml:"schema"`
	Tables []string `yaml:"tables"`
}
