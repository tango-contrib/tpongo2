package tpongo2

import (
	"testing"
	"bytes"
	"net/http"
	"reflect"
	"net/http/httptest"

	"github.com/flosch/pongo2"
	"github.com/lunny/tango"
)

type RenderAction struct {
	Renderer
}

func (a *RenderAction) Get() error {
	return a.RenderString("Hello {{ name }}!", pongo2.Context{
		"name": "tango",
	})
}

func TestPango2_1(t *testing.T) {
	buff := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	recorder.Body = buff

	o := tango.Classic()
	o.Use(New())
	o.Get("/", new(RenderAction))

	req, err := http.NewRequest("GET", "http://localhost:3000/", nil)
	if err != nil {
		t.Error(err)
	}

	o.ServeHTTP(recorder, req)
	expect(t, recorder.Code, http.StatusOK)
	refute(t, len(buff.String()), 0)
	expect(t, buff.String(), "Hello tango!")
}

type Render2Action struct {
	Renderer
}

func (a *Render2Action) Get() error {
	return a.Render("test1.html", pongo2.Context{
		"name": "tango",
	})
}

func TestPango2_2(t *testing.T) {
	buff := bytes.NewBufferString("")
	recorder := httptest.NewRecorder()
	recorder.Body = buff

	o := tango.Classic()
	o.Use(New())
	o.Get("/", new(Render2Action))

	req, err := http.NewRequest("GET", "http://localhost:3000/", nil)
	if err != nil {
		t.Error(err)
	}

	o.ServeHTTP(recorder, req)
	expect(t, recorder.Code, http.StatusOK)
	refute(t, len(buff.String()), 0)
	expect(t, buff.String(), "Hello tango!")
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}