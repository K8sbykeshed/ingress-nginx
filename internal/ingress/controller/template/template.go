package template

import "k8s.io/ingress-nginx/internal/ingress/controller/config"

// Writer is the interface to render a template
type Writer interface {
	// Write renders the template.
	// NOTE: Implementers must ensure that the content of the returned slice is not modified by the implementation
	// after the return of this function.
	Write(conf *config.TemplateConfig) ([]byte, error)
	// Validate is a function that can be called, containing the file name to be tested
	// This function should be used just by specific cases like crossplane, otherwise it can return
	// null error
	Validate(filename string) error
}

type LuaConfig struct {
	EnableMetrics           bool           `json:"enable_metrics"`
	ListenPorts             LuaListenPorts `json:"listen_ports"`
	UseForwardedHeaders     bool           `json:"use_forwarded_headers"`
	UseProxyProtocol        bool           `json:"use_proxy_protocol"`
	IsSSLPassthroughEnabled bool           `json:"is_ssl_passthrough_enabled"`
	HTTPRedirectCode        int            `json:"http_redirect_code"`
	EnableOCSP              bool           `json:"enable_ocsp"`
	MonitorBatchMaxSize     int            `json:"monitor_batch_max_size"`
	HSTS                    bool           `json:"hsts"`
	HSTSMaxAge              string         `json:"hsts_max_age"`
	HSTSIncludeSubdomains   bool           `json:"hsts_include_subdomains"`
	HSTSPreload             bool           `json:"hsts_preload"`
}

type LuaListenPorts struct {
	HTTPSPort    string `json:"https"`
	StatusPort   string `json:"status_port"`
	SSLProxyPort string `json:"ssl_proxy"`
}
