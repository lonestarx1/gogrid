// Package templates provides project scaffolding templates for gogrid init.
package templates

// Data holds the template rendering context.
type Data struct {
	// Name is the project name.
	Name string
	// Module is the Go module path (e.g. "github.com/example/myproject").
	Module string
}

// File represents a single file to generate.
type File struct {
	// Path is the relative path within the project directory.
	Path string
	// Content is the Go text/template string.
	Content string
}

// Template defines a complete project scaffold.
type Template struct {
	// Name is the template identifier (e.g. "single", "team", "pipeline").
	Name string
	// Description is a short description.
	Description string
	// Files are the files to generate.
	Files []File
}

var registry = map[string]*Template{}

func register(t *Template) {
	registry[t.Name] = t
}

// Get returns the template with the given name, or nil.
func Get(name string) *Template {
	return registry[name]
}
