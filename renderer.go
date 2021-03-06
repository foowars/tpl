package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// Renderer will render a set of inputs.
type Renderer struct {
	FuncMap      template.FuncMap
	Inputs       []string
	PreloadFiles []string
	StopOnError  bool
}

// Execute applies a dataset against all inputs and writes output.
func (r *Renderer) Execute(out string, values map[string]interface{}) error {
	return r.execute(r.Inputs, out, values)
}

func (r *Renderer) execute(inputs []string, out string, values map[string]interface{}) error {
	// Do not order inputs, which may have been provided in a specific order
	// from the command line
	for _, fn := range inputs {
		f, err := os.Open(fn)
		if err != nil {
			return err
		}

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		// Render files directly
		if !fi.IsDir() {
			withPreloads := make([]string, 0)
			for _, lib := range r.PreloadFiles {
				withPreloads = append(withPreloads, lib)
			}
			withPreloads = append(withPreloads, fn)

			err := r.render(values, withPreloads, r.getOutputPath(out, path.Base(fn)))
			if err != nil {
				return err
			}
			continue
		}

		// Loop through each file in a directory and render it
		eis, err := f.Readdirnames(-1)
		if err != nil {
			return err
		}

		// Pluck out absolute path names; unlike inputs, these are safe to sort,
		// because they were generated values
		names := stringSorter{}
		for _, ei := range eis {
			names = append(names, filepath.Join(f.Name(), ei))
		}
		sort.Sort(names)

		outpath := out
		if strings.HasSuffix(out, "/") {
			outpath = outpath + path.Base(f.Name()) + "/"
		}

		err = r.execute(names, outpath, values)
		if err != nil {
			return err
		}

	}
	return nil
}

func (r *Renderer) getOutputPath(base, fn string) string {
	if base == "" || base == "-" {
		return "-"
	}
	if strings.HasSuffix(fn, ".tpl") {
		fn = strings.TrimSuffix(fn, ".tpl")
	} else if strings.HasSuffix(fn, ".tmpl") {
		fn = strings.TrimSuffix(fn, ".tmpl")
	}
	if strings.HasSuffix(base, "/") {
		return filepath.Join(base, fn)
	}
	if f, err := os.Open(base); err == nil {
		if fi, err := f.Stat(); err == nil {
			if fi.IsDir() {
				return filepath.Join(base, fn)
			}
		}
	}
	return base
}

func (r *Renderer) render(values map[string]interface{}, inames []string, oname string) error {
	if oname == "" {
		return errors.New("Output name cannot be blank")
	}

	var out *os.File
	var err error
	if oname == "-" {
		out = os.Stdout
		log.Printf("Rendering [%s] to STDOUT\n", strings.Join(inames, ", "))
	} else {
		if strings.Contains(oname, "/") {
			if err := os.MkdirAll(path.Dir(oname), 0755); err != nil {
				return fmt.Errorf("Error creating directory for %q: %v", oname, err)
			}
		}

		out, err = os.OpenFile(oname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("Cannot open output file %q: %v", oname, err)
		}

		log.Printf("Rendering [%s] into %s\n", strings.Join(inames, ", "), oname)
		defer func() { out.Sync(); out.Close() }()
	}

	tpl := template.New(filepath.Base(inames[len(inames)-1]))
	if r.FuncMap != nil {
		tpl.Funcs(r.FuncMap)
	}

	_, err = tpl.ParseFiles(inames...)
	if err != nil {
		return fmt.Errorf("Cannot parse templates [%s]: %v", strings.Join(inames, ", "), err)
	}

	if r.StopOnError {
		tpl.Option("missingkey=error")
	} else {
		tpl.Option("missingkey=zero")
	}

	return tpl.Execute(out, values)
}
