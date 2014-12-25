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

type Pangoer interface {
	SetRenderer(renderer *Renderer)
}

type Render struct {
	*Renderer
}

func (r *Render) SetRenderer(renderer *Renderer) {
	r.Renderer = renderer
}

type Pongo struct {
	log tango.Logger
	templatesDir string
	templates map[string]*pongo2.Template
	lock sync.RWMutex
	moniter bool
}

func New(templatesDir string, moniter bool) *Pongo {
	return &Pongo{
		templatesDir: templatesDir,
		templates: make(map[string]*pongo2.Template),
		moniter: moniter,
	}
}

func Default() *Pongo {
	return New("templates", false)
}

// @inject
func (p *Pongo) SetLogger(l tango.Logger) {
	p.log = l
}

func (p *Pongo) GetTemplate(name string) (t *pongo2.Template, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var ok bool
	if t, ok = p.templates[name]; !ok {
		t, err = pongo2.FromFile(filepath.Join(p.templatesDir, name))
		if err != nil {
			return
		}
		p.templates[name] = t
	}
	return
}

func (p *Pongo) NewRenderer(resp tango.ResponseWriter) *Renderer {
	return &Renderer{
		render: p,
		templatesDir: p.templatesDir,
		ResponseWriter: resp,
		ContentType: ContentHTML,
		Charset: DefaultCharset,
	}
}

func (p *Pongo) Handle(ctx *tango.Context) {
	if action := ctx.Action(); action != nil {
		if pr, ok := action.(Pangoer); ok {
			rd := p.NewRenderer(ctx.ResponseWriter)
			pr.SetRenderer(rd)
		}
	}
	ctx.Next()
}

type Renderer struct {
	render *Pongo
	templatesDir string

	tango.ResponseWriter
	ContentType string
	Charset string
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