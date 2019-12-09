package gen

import (
	"bytes"
	"path"
	"path/filepath"

	"github.com/nanitefactory/chromebot-domain-gen/gen/gotpl"
	"github.com/valyala/quicktemplate"

	"github.com/chromedp/cdproto-gen/gen/genutil"
	"github.com/chromedp/cdproto-gen/pdl"
)

// Generator is the common interface for code generators.
type Generator func([]*pdl.Domain, string) (Emitter, error)

// Emitter is the shared interface for code emitters.
type Emitter interface {
	Emit() map[string]*bytes.Buffer
}

// Generators returns all the various Chrome DevTools Protocol generators.
func Generators() map[string]Generator {
	return map[string]Generator{
		"go": NewGoGenerator,
	}
}

// GoGenerator generates Go source code for the Chrome DevTools Protocol.
type GoGenerator struct {
	files fileBuffers
}

// NewGoGenerator creates a Go source code generator for the Chrome DevTools
// Protocol domain definitions.
func NewGoGenerator(domains []*pdl.Domain, outBasePkg string) (Emitter, error) {
	var w *quicktemplate.Writer

	fb := make(fileBuffers)

	// generate individual domains
	for _, d := range domains {
		pkgName := genutil.PackageName(d)
		pkgOut := filepath.Join(pkgName, pkgName+".go")

		// do command template
		w = fb.get(pkgOut, pkgName, d, domains, "github.com/chromedp/cdproto")
		gotpl.StreamDomainTemplate(w, d, domains)
		fb.release(w)
	}

	return &GoGenerator{
		files: fb,
	}, nil
}

// Emit returns the generated files.
func (gg *GoGenerator) Emit() map[string]*bytes.Buffer {
	return map[string]*bytes.Buffer(gg.files)
}

// fileBuffers is a type to manage buffers for file data.
type fileBuffers map[string]*bytes.Buffer

// get retrieves the file buffer for s, or creates it if it is not yet available.
func (fb fileBuffers) get(s string, pkgName string, d *pdl.Domain, domains []*pdl.Domain, cdprotoBasePkg string) *quicktemplate.Writer {
	// check if it already exists
	if b, ok := fb[s]; ok {
		return quicktemplate.AcquireWriter(b)
	}

	// create buffer
	b := new(bytes.Buffer)
	fb[s] = b
	w := quicktemplate.AcquireWriter(b)

	v := d
	if b := path.Base(s); b != pkgName+".go" {
		v = nil
	}

	// add package header
	gotpl.StreamFileHeader(w, "domain", v) // "package domain" for all domains

	// add import map
	importMap := map[string]string{
		"context":                            "",
		"encoding/json":                      "",
		cdprotoBasePkg + "/cdp":              "",
		"github.com/mailru/easyjson":         "",
		"github.com/mailru/easyjson/jlexer":  "",
		"github.com/mailru/easyjson/jwriter": "",
	}
	for _, d := range domains {
		importMap[cdprotoBasePkg+"/"+genutil.PackageName(d)] = ""
	}
	gotpl.StreamFileImportTemplate(w, importMap)

	return w
}

// release releases a template writer.
func (fb fileBuffers) release(w *quicktemplate.Writer) {
	quicktemplate.ReleaseWriter(w)
}
