package dsl

// Bundle is a rules file export manifest.
type Bundle struct {
	// Version is a bundle version.
	// It's preferred to use a semver format.
	// Examples: "0.5.1", "1.0.0".
	Version string

	// TODO: what else do we need here?
}

// ImportRules imports all rules from the bundle and prefixes them with a specified string.
// Only packages that have an exported Bundle variable can be imported.
func ImportRules(prefix string, bundle Bundle) {}
