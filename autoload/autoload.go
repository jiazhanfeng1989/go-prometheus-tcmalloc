// Package autoload registers the tcmalloc collector on the default Prometheus registry.
//
// Blank-import it from main. Do not also call tcmallocprom.NewCollector on that registry.
//
//	import _ "github.com/jiazhanfeng1989/go-prometheus-tcmalloc/autoload"
package autoload

import (
	tcmallocprom "github.com/jiazhanfeng1989/go-prometheus-tcmalloc"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(tcmallocprom.NewCollector())
}
