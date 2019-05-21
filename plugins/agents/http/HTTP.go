package http

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gansoi/gansoi/build"
	"github.com/gansoi/gansoi/plugins"
)

func init() {
	plugins.RegisterAgent("http", HTTP{})
}

const (
	redirectsToFollow = 10

	httpsScheme = "https"
	httpScheme  = "http"
)

var (
	dial = net.Dial

	userAgent = build.UserAgent + " http-agent"
)

// HTTP will request a ressource from a HTTP server.
type HTTP struct {
	URL            string `json:"url" description:"The URL to request"`
	FollowRedirect bool   `json:"followRedirect" description:"Follow 30x redirects" default:"true"`
	Insecure       bool   `json:"insecure" description:"Ignore SSL errors"`
	IncludeBody    bool   `json:"includeBody" description:"Include body in results"`
	Host           string `json:"host" description:"Host to contact (leave empty to use host derived from URL)"`
	method         string // This is only used for testing (for now), NewRequest will default to "GET".
}

func getHostPort(URL *url.URL) (string, string) {
	if !strings.ContainsRune(URL.Host, ':') {
		switch URL.Scheme {
		case httpScheme:
			return URL.Host, "80"
		case httpsScheme:
			return URL.Host, "443"
		}
	}

	host, port, _ := net.SplitHostPort(URL.Host)

	return host, port
}

func camelCaseHeader(key string) string {
	fields := strings.FieldsFunc(key, func(r rune) bool {
		return !plugins.ValidateResultKeyRune(r)
	})

	result := "Header"

	for _, f := range fields {
		result += strings.Title(f)
	}

	return result
}

// Check implements plugins.Agent.
func (h *HTTP) Check(result plugins.AgentResult) error {
	URL, err := url.Parse(h.URL)
	if err != nil {
		return err
	}

	for try := 0; try < redirectsToFollow; try++ {
		if !(URL.Scheme == "http" || URL.Scheme == "https") {
			return http.ErrNotSupported
		}

		host, port := getHostPort(URL)

		if h.Host != "" {
			host = h.Host
		}

		t0 := time.Now()
		raddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
		if err != nil {
			return err
		}

		t1 := time.Now()
		var conn net.Conn
		conn, err = dial("tcp", raddr.String())
		if err != nil {
			return err
		}
		defer conn.Close()

		t2 := time.Now()
		if URL.Scheme == httpsScheme {
			c := tls.Client(conn, &tls.Config{
				ServerName:         host,
				InsecureSkipVerify: h.Insecure,
			})

			err = c.Handshake()
			if err != nil {
				return err
			}

			state := c.ConnectionState()

			if len(state.PeerCertificates) > 0 {
				cert := state.PeerCertificates[0]
				notAfter := cert.NotAfter
				result.AddValue("SSLValidDays", time.Until(notAfter).Hours()/24.0)
				result.AddValue("SSLCommonName", cert.Subject.CommonName)
			}

			conn = c
		}

		req, err := http.NewRequest(h.method, URL.String(), nil)
		if err != nil {
			return err
		}

		req.Header.Add("User-Agent", userAgent)

		t3 := time.Now()
		err = req.Write(conn)
		if err != nil {
			return err
		}

		t4 := time.Now()
		resp, err := http.ReadResponse(bufio.NewReader(conn), req)
		if err != nil {
			return err
		}

		if h.FollowRedirect && (resp.StatusCode == 301 || resp.StatusCode == 302) {
			URL, err = resp.Location()
			if err != nil {
				return err
			}

			result.AddValue(fmt.Sprintf("Redirect%d", try), URL.String())

			continue
		}

		t5 := time.Now()
		b := make([]byte, 1024)
		n, _ := resp.Body.Read(b)
		resp.Body.Close()
		t6 := time.Now()

		if h.IncludeBody {
			result.AddValue("Body", string(b[:n]))
		}

		result.AddValue("StatusCode", resp.StatusCode)
		result.AddValue("TimeDNS", ms(t1.Sub(t0)))
		result.AddValue("TimeConnect", ms(t2.Sub(t1)))
		result.AddValue("TimeTLS", ms(t3.Sub(t2)))
		result.AddValue("TimeRequest", ms(t4.Sub(t3)))
		result.AddValue("TimeReadHeaders", ms(t5.Sub(t4)))
		result.AddValue("TimeReadBody", ms(t6.Sub(t5)))
		result.AddValue("TimeAccumulated", ms(t6.Sub(t0)))

		for k, v := range resp.Header {
			result.AddValue(camelCaseHeader(k), strings.Join(v, " "))
		}

		break
	}

	return nil
}

// ms will convert a time.Duration to milliseconds.
func ms(d time.Duration) int64 {
	return (d.Nanoseconds() + 1000000/2) / 1000000
}
