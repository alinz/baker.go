package baker

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Processor interface {
	Request(r *http.Request)
	Response(r *http.Response) error
	HandleError(rw http.ResponseWriter, r *http.Request, err error)
}

//
// PathAppend
//

type ProcessorPathAppend struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

var _ Processor = (*ProcessorPathAppend)(nil)

func (p *ProcessorPathAppend) Request(r *http.Request) {
	var builder strings.Builder

	if p.Begin != "" {
		builder.WriteString(p.Begin)
	}

	builder.WriteString(r.URL.Path)

	if p.End != "" {
		builder.WriteString(p.End)
	}

	r.URL.Path = builder.String()
}

func (p *ProcessorPathAppend) Response(r *http.Response) error {
	return nil
}

func (p *ProcessorPathAppend) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

func CreateProcessorPathAppend(raw json.RawMessage) (Processor, error) {
	r := &ProcessorPathAppend{}

	err := json.Unmarshal(raw, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//
// PathReplace
//

type ProcessorPathReplace struct {
	Search  string `json:"search"`
	Replace string `json:"replace"`
	Times   int    `json:"times"`
}

var _ Processor = (*ProcessorPathReplace)(nil)

func (p *ProcessorPathReplace) Request(r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, p.Search, p.Replace, p.Times)
}

func (p *ProcessorPathReplace) Response(r *http.Response) error {
	return nil
}

func (p *ProcessorPathReplace) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}

func CreateProcessorPathReplace(raw json.RawMessage) (Processor, error) {
	r := &ProcessorPathReplace{}

	err := json.Unmarshal(raw, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}
