package main

import (
	"bytes"
	"container/ring"
	"fmt"
	"net/http"
	"text/template"

	"github.com/yosisa/fluxion/event"
	"github.com/yosisa/fluxion/plugin"
)

type Config struct {
	Bind     string `codec:"bind"`
	Template string `codec:"template"`
	Row      int    `codec:"row"`
}

type StatusBoardFilter struct {
	conf *Config
	tmpl *template.Template
	rows *ring.Ring
}

func (f *StatusBoardFilter) Init(fn plugin.ConfigFeeder) (err error) {
	f.conf = &Config{}
	if err = fn(f.conf); err != nil {
		return
	}
	if f.tmpl, err = template.New("").Parse(f.conf.Template); err != nil {
		return
	}
	if f.conf.Row == 0 {
		f.conf.Row = 4
	}
	f.rows = ring.New(f.conf.Row)
	return
}

func (f *StatusBoardFilter) Start() error {
	http.Handle("/", f)
	go http.ListenAndServe(f.conf.Bind, nil)
	plugin.Log.Info("Server started on ", f.conf.Bind)
	return nil
}

func (f *StatusBoardFilter) Filter(r *event.Record) (*event.Record, error) {
	w := &bytes.Buffer{}
	f.tmpl.Execute(w, r)
	f.rows.Value = w.String()
	f.rows = f.rows.Next()
	return r, nil
}

func (f *StatusBoardFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<table>\n")
	e := f.rows.Prev()
	for i := 0; i < f.conf.Row && e.Value != nil; i++ {
		fmt.Fprintf(w, "<tr><td>%v</td></tr>\n", e.Value)
		e = e.Prev()
	}
	fmt.Fprint(w, "</table>\n")
}

func main() {
	plugin.New(func() plugin.Plugin {
		return &StatusBoardFilter{}
	}).Run()
}
