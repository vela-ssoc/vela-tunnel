package tunnel

import "testing"

type hostport struct {
	Host string
	Port string
}

func TestURL(t *testing.T) {
	qas := map[string]hostport{
		"wzy.example.com:443":         {Host: "wzy.example.com", Port: "443"},
		"tls://wzy.example.com:443":   {Host: "wzy.example.com", Port: "443"},
		"https://wzy.example.com:443": {Host: "wzy.example.com", Port: "443"},
		"https://wzy.example.com":     {Host: "wzy.example.com", Port: ""},
		"1.2.3.4:443":                 {Host: "1.2.3.4", Port: "443"},
		"tls://1.2.3.4:443":           {Host: "1.2.3.4", Port: "443"},
		"https://1.2.3.4:443":         {Host: "1.2.3.4", Port: "443"},
		"[::1]:443":                   {Host: "::1", Port: "443"},
		"https://[::1]":               {Host: "::1", Port: ""},
		"https://[::1]:443":           {Host: "::1", Port: "443"},
	}

	for q, a := range qas {
		host, port := splitHostPort(q)
		if host != a.Host || port != a.Port {
			t.Errorf("[xxx] %s -> host: %s, port: %s, expected host: %s, port: %s", q, host, port, a.Host, a.Port)
		} else {
			t.Logf("%s -> host: %s, port: %s", q, host, port)
		}
	}
}
