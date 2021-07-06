package cli

type Flag struct {
	Name     string   // --help
	Aliases  []string // -h
	Usage    string   //
	Default  string   //
	Required bool     //
	Hidden   bool     //
}
