package tpongo2

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/flosch/pongo2"
	"github.com/lunny/tango"
)

const (
	ContentType           = "Content-Type"
	ContentLength         = "Content-Length"
	ContentBinary         = "application/octet-stream"
	ContentJSON           = "application/json"
	ContentHTML           = "text/html"
	ContentXHTML          = "application/xhtml+xml"
	ContentXML            = "text/xml"
	DefaultCharset        = "UTF-8"
	DefaultTemplateSuffix = ".html"
)

type Pongoer interface {
	SetRenderer(*Pongo, tango.ResponseWriter, string, string)
}

type Renderer struct {
	render *Pongo
	tango.ResponseWriter
	ContentType string
	Charset     string
}

func (r *Renderer) SetRenderer(render *Pongo, resp tango.ResponseWriter,
	ContentType, Charset string) {
	r.render = render
	r.ResponseWriter = resp
	r.ContentType = ContentType
	r.Charset = Charset
}

type Pongo struct {
	Options

	templates map[string]*pongo2.Template
	lock      sync.RWMutex
}

type Options struct {
	TemplatesDir string
	Reload       bool
	Suffix       string
}

func New(opts ...Options) *Pongo {
	opt := prepareOptions(opts)
	return &Pongo{
		Options:   opt,
		templates: make(map[string]*pongo2.Template),
	}
}

func Default() *Pongo {
	return New()
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	if opt.TemplatesDir == "" {
		opt.TemplatesDir = "templates"
	}
	if len(opt.Suffix) <= 0 {
		opt.Suffix = DefaultTemplateSuffix
	}

	if opt.Suffix[0] != '.' {
		opt.Suffix = "." + opt.Suffix
	}

	return opt
}

func (p *Pongo) GetTemplate(name string) (t *pongo2.Template, err error) {
	if !strings.HasSuffix(name, p.Suffix) {
		name = name + p.Suffix
	}
	if p.Reload {
		return pongo2.FromFile(filepath.Join(p.Options.TemplatesDir, name))
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	var ok bool
	if t, ok = p.templates[name]; !ok {
		t, err = pongo2.FromFile(filepath.Join(p.Options.TemplatesDir, name))
		if err != nil {
			return
		}
		p.templates[name] = t
	}
	return
}

func (p *Pongo) Handle(ctx *tango.Context) {
	if action := ctx.Action(); action != nil {
		if pr, ok := action.(Pongoer); ok {
			pr.SetRenderer(p, ctx.ResponseWriter, ContentHTML, DefaultCharset)
		}
	}
	ctx.Next()
}

type T map[string]interface{}

func (r *Renderer) Render(tmpl string, data map[string]interface{}) error {
	return r.RenderFile(tmpl, pongo2.Context(data))
}

func (r *Renderer) RenderFile(tmpl string, data map[string]interface{}) error {
	t, err := r.render.GetTemplate(tmpl)
	if err != nil {
		return err
	}

	r.Header().Set(ContentType, r.ContentType+"; charset="+r.Charset)
	if err := t.ExecuteWriter(data, r.ResponseWriter); err != nil {
		return err
	}
	return nil
}

// TODO: should add cache
func (r *Renderer) RenderString(content string, data pongo2.Context) error {
	tpl, err := pongo2.FromString(content)
	if err != nil {
		return err
	}

	r.Header().Set(ContentType, r.ContentType+"; charset="+r.Charset)
	return tpl.ExecuteWriter(data, r.ResponseWriter)
}

func (r *Renderer) HTMLBytes(tmpl string, data map[string]interface{}) ([]byte, error) {
	t, err := r.render.GetTemplate(tmpl)
	if err != nil {
		return nil, err
	}

	r.Header().Set(ContentType, r.ContentType+"; charset="+r.Charset)
	return t.ExecuteBytes(data)

}

func (r *Renderer) HTMLString(tmpl string, data map[string]interface{}) (string, error) {
	b, e := r.HTMLBytes(tmpl, data)
	return string(b), e
}
