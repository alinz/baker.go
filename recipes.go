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
