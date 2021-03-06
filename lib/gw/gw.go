package gw

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/digitalcircle-com-br/caroot"

	"github.com/digitalcircle-com-br/compass/lib/mw"
	"github.com/digitalcircle-com-br/compass/lib/types"
	"github.com/digitalcircle-com-br/service"
	"golang.org/x/crypto/acme/autocert"
)

var cli = http.Client{}
var cfg *types.Config
var dbg = func(s string, i ...interface{}) {

}

func tpcPingHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("addr")
	if addr == "" {
		http.Error(w, "addr not provided", http.StatusBadRequest)
		return
	}
	start := time.Now()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	conn.Close()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Connection ok - time: %s", time.Since(start).String())))
}

func curlHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("url")
	method := r.URL.Query().Get("method")
	if method == "" {
		method = http.MethodGet
	}
	defer r.Body.Close()

	req, err := http.NewRequest(method, addr, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Header = r.Header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()
	for k, v := range res.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func buildRoutes(cfg *types.Config, route *types.Route) http.HandlerFunc {
	switch route.Target {
	case "@ping":
		return tpcPingHandler
	case "@curl":
		return curlHandler
	default:
		return buildRoute(cfg, route)
	}
}

func buildRoute(cfg *types.Config, route *types.Route) http.HandlerFunc {

	allhops := make([]string, 0)

	allhops = append(allhops, cfg.Hops...)

	allhops = append(allhops, route.Hops...)

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fhost := r.Host
		npath := strings.Replace(r.URL.Path, route.Path, route.Target, 1)
		service.Debug("Replacing: %s - %s  - %s => %s", r.URL.Path, route.Path, route.Target, npath)
		queryParams := r.URL.Query().Encode()
		if len(queryParams) > 0 {
			npath = npath + "?" + queryParams
		}

		freq, err := http.NewRequest(r.Method, npath, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		freq.Header = r.Header
		for i := 0; i < len(allhops); i++ {

			req, err := http.NewRequest(http.MethodGet, allhops[i], nil)
			req.Header = freq.Header
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			res, err := cli.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if res.StatusCode != http.StatusOK {
				w.WriteHeader(res.StatusCode)
				return
			}
			for k, v := range res.Header {
				for _, vv := range v {
					freq.Header.Add(k, vv)
				}
			}
			service.ServerTiming(w, fmt.Sprintf("mw-%v", 1), allhops[i], start)

		}

		dbg("Forwarding %s to %s", r.URL.String(), freq.URL.String())

		res, err := cli.Do(freq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for k, v := range res.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.Header().Set("Host", fhost)
		w.WriteHeader(res.StatusCode)
		service.ServerTiming(w, "endpoint", "Endpoint", start)
		defer res.Body.Close()
		io.Copy(w, res.Body)

	}
}

var hostMuxes = make(map[string]*http.ServeMux)

func Setup() error {
	running := false
	cfg = &types.Config{}
	ch := service.Config(cfg)
	for {
		_, ok := <-ch

		if cfg.Debug {
			dbg = func(s string, i ...interface{}) {
				service.Debug(s, i...)
			}
		} else {
			dbg = func(s string, i ...interface{}) {

			}
		}

		if !ok {
			return nil
		}
		hostMuxes = make(map[string]*http.ServeMux)
		for k := range cfg.Hosts {
			host := cfg.Hosts[k]

			service.Log("Setting up hosts: %s", host.Host)
			h, ok := hostMuxes[host.Host]
			if !ok {
				h = http.NewServeMux()
				hostMuxes[host.Host] = h
			}
			for _, r := range host.Routes {
				h.HandleFunc(r.Path, buildRoutes(cfg, r))
			}
		}
		if !running {
			go run()
			running = true
		}
	}
}

func run() error {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		k := strings.Split(r.Host, ":")[0]
		m, ok := hostMuxes[k]
		if ok {
			dbg("Serving route: %s => %s", k, r.URL.String())
			m.ServeHTTP(w, r)
			return
		}
		m, ok = hostMuxes["*"]
		if ok {
			dbg("Serving default: %s => %s", k, r.URL.String())
			m.ServeHTTP(w, r)
			return
		}
		dbg("No route found for: %s", r.URL.String())
		http.NotFound(w, r)

	}
	h = mw.CSP("", h)
	h = mw.STS("", h)
	h = mw.XFrame("", h)
	h = mw.Helmet(h)
	h = mw.CORS(h)

	http.HandleFunc("/", h)

	if !cfg.Acme {
		if cfg.SelfSigned && cfg.Https {
			service.Log("Using https + self signed approach")
			tlscfg := &tls.Config{
				GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
					ca := caroot.GetOrGenFromRoot(info.ServerName)
					return ca, nil
				},
			}

			server := &http.Server{
				Addr:      cfg.Addr,
				Handler:   http.DefaultServeMux,
				TLSConfig: tlscfg,
			}
			go func() {
				err := server.ListenAndServeTLS("", "")
				if err != nil {
					service.Log("Finishing server: %s", err.Error())
				}
			}()

		} else {

			if cfg.Addr == "" {
				cfg.Addr = ":8080"
			}

			service.Log("No acme set - if required, set ENV VAR ACME")

			if cfg.Https {
				service.Log("Using simple ssl approach")
				conn, err := net.Listen("tcp", cfg.Addr)
				if err != nil {
					return err
				}
				server := http.Server{}
				service.Log("APIGW - Running HTTPS @ %v", cfg.Addr)
				go func() {
					err = server.ServeTLS(conn, cfg.Cert, cfg.Key)
					if err != nil {
						service.Fatal(err)
					}
				}()
			} else {
				service.Log("Using simple approach")
				server := &http.Server{Addr: cfg.Addr}
				service.Log("APIGW - Running HTTP @ %v", cfg.Addr)
				go func() {
					err := server.ListenAndServe()
					if err != nil {
						service.Fatal(err)
					}
				}()
			}
		}

	} else {
		service.Log("Using ACME approach")
		if cfg.Certs == "" {
			cfg.Certs = "./certs"
		}
		service.Log("ACME Found!")
		certManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(path.Join("./certs")),
		}

		server := &http.Server{
			Addr: ":443",

			TLSConfig: &tls.Config{
				PreferServerCipherSuites: true,
				// Only use curves which have assembly implementations
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519, // Go 1.8 only
				},
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

					// Best disabled, as they don't provide Forward Secrecy,
					// but might be necessary for some clients
					// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				},
				GetCertificate:     certManager.GetCertificate,
				InsecureSkipVerify: true,
			},
		}

		go func() {
			err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
			if err != nil {
				service.Fatal(err)
			}
		}()
		go func() {
			err := server.ListenAndServeTLS("", "")
			if err != nil {
				service.Fatal(err)
			}
		}()
	}

	for {
		time.Sleep(time.Minute)
	}

}
