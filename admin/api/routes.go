package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/fabiolb/fabio/route"
)

type RoutesHandler struct{}

type apiRoute struct {
	Service string   `json:"service"`
	Host    string   `json:"host"`
	Path    string   `json:"path"`
	Src     string   `json:"src"`
	Dst     string   `json:"dst"`
	Opts    string   `json:"opts"`
	Weight  float64  `json:"weight"`
	Tags    []string `json:"tags,omitempty"`
	Cmd     string   `json:"cmd"`
	Rate1   float64  `json:"rate1"`
	Pct99   float64  `json:"pct99"`
}

func (h *RoutesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := route.GetTable()

	if _, ok := r.URL.Query()["raw"]; ok {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, t.String())
		return
	}

	var hostFilter *string
	var pathFilter *string
	var tagFilter *string

	if param, ok := r.URL.Query()["host"]; ok {
		hostFilter = &param[0]
	}

	if param, ok := r.URL.Query()["path"]; ok {
		pathFilter = &param[0]
	}

	if param, ok := r.URL.Query()["tag"]; ok {
		tagFilter = &param[0]
	}

	var hosts []string

	if hostFilter != nil {
		hosts = append(hosts, *hostFilter)
	} else {
		for host := range t {
			hosts = append(hosts, host)
		}
		sort.Strings(hosts)
	}

	var routes []apiRoute

	for _, host := range hosts {
		for _, tr := range t[host] {
			if pathFilter == nil || tr.Path == *pathFilter {
				for _, tg := range tr.Targets {

					if tagFilter == nil || stringInSlice(*tagFilter, tg.Tags) {
						var opts []string
						for k, v := range tg.Opts {
							opts = append(opts, k+"="+v)
						}

						ar := apiRoute{
							Service: tg.Service,
							Host:    tr.Host,
							Path:    tr.Path,
							Src:     tr.Host + tr.Path,
							Dst:     tg.URL.String(),
							Opts:    strings.Join(opts, " "),
							Weight:  tg.Weight,
							Tags:    tg.Tags,
							Cmd:     "route add",
							Rate1:   tg.Timer.Rate1(),
							Pct99:   tg.Timer.Percentile(0.99),
						}
						routes = append(routes, ar)
					}

				}
			}
		}
	}
	writeJSON(w, r, routes)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
