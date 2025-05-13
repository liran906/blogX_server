package flags

import "flag"

type Options struct {
	File    string
	DB      bool
	Version bool
}

var FlagOptions = new(Options)

func Parse() {
	flag.StringVar(&FlagOptions.File, "f", "settings.yaml", "config file")
	flag.BoolVar(&FlagOptions.DB, "db", false, "database migration")
	flag.BoolVar(&FlagOptions.Version, "v", false, "show version")
	flag.Parse()
}
