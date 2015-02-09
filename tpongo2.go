package tpongo2

import (
	"path/filepath"
	"sync"

	"github.com/lunny/tango"
	"gopkg.in/flosch/pongo2.v3"
)

const (
	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
	ContentBinary  = "application/octet-stream"
	ContentJSON    = "application/json"
	ContentHTML    = "text/html"
	ContentXHTML   = "application/xhtml+xml"
	ContentXML     = "text/xml"
	DefaultCharset = "UTF-8"
)

type Pongoer interface {
	SetRenderer(*renderer)
}

type Renderer struct {
	*renderer
}

func (r *Renderer) SetRenderer(renderer *renderer) {
	r.renderer = renderer
}

type Pongo struct {
	Options

	templates map[string]*pongo2.Template
	lock sync.RWMutex
}

type Options struct {
	TemplatesDir string
	Reload bool
}

func New(opts ...Options) *Pongo {
	opt := prepareOptions(opts)
	return &Pongo{
		Options : opt,
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
	return opt
}

func (p *Pongo) GetTemplate(name string) (t *pongo2.Template, err error) {
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

func (p *Pongo) NewRenderer(resp tango.ResponseWriter) *renderer {
	return &renderer{
		render: p,
		ResponseWriter: resp,
		ContentType: ContentHTML,
		Charset: DefaultCharset,
	}
}

func (p *Pongo) Handle(ctx *tango.Context) {
	if action := ctx.Action(); action != nil {
		if pr, ok := action.(Pongoer); ok {
			rd := p.NewRenderer(ctx.ResponseWriter)
			pr.SetRenderer(rd)
		}
	}
	ctx.Next()
}

type renderer struct {
	render *Pongo

	tango.ResponseWriter
	ContentType string
	Charset string
}

func (r *Renderer) Render(tmpl string, data pongo2.Context) error {
	return r.RenderFile(tmpl, data)
}

func (r *Renderer) RenderFile(tmpl string, data pongo2.Context) error {
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