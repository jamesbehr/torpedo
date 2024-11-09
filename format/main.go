package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Formatter interface {
	Write(map[string]any) error
	Close() error
}

func New(format string, fields []string, w io.Writer) (Formatter, error) {
	switch format {
	case "json":
		return &jsonFormatter{fields, json.NewEncoder(w)}, nil
	case "text":
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)
		return &textFormatter{false, fields, tw}, nil
	default:
		return nil, fmt.Errorf("invalid format %q", format)
	}
}

type textFormatter struct {
	wroteHeaders bool
	fields       []string
	tw           *tabwriter.Writer
}

func (f *textFormatter) Write(m map[string]any) error {
	if f.fields == nil {
		for key := range m {
			f.fields = append(f.fields, key)
		}
	}

	if !f.wroteHeaders {
		if _, err := fmt.Fprintln(f.tw, strings.Join(f.fields, "\t")); err != nil {
			return err
		}

		f.wroteHeaders = true
	}

	row := make([]string, len(f.fields))
	for i, field := range f.fields {
		row[i] = fmt.Sprint(m[field])
	}

	if _, err := fmt.Fprintln(f.tw, strings.Join(row, "\t")); err != nil {
		return err
	}

	return nil
}

func (f *textFormatter) Close() error { return f.tw.Flush() }

type jsonFormatter struct {
	fields []string
	*json.Encoder
}

func (f *jsonFormatter) Write(m map[string]any) error {
	if f.fields == nil {
		for key := range m {
			f.fields = append(f.fields, key)
		}
	}

	filtered := map[string]any{}
	for _, field := range f.fields {
		filtered[field] = m[field]
	}

	return f.Encode(filtered)
}

func (f *jsonFormatter) Close() error { return nil }
