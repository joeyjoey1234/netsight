package tools

import (
	"context"
	"fmt"
	"net"
	"strings"
)

type DNSResult struct {
	QueryType string   `json:"queryType"`
	Query     string   `json:"query"`
	Results   []string `json:"results"`
	TTL       uint32   `json:"ttl,omitempty"`
	Server    string   `json:"server"`
	Error     string   `json:"error,omitempty"`
}

func NSLookup(ctx context.Context, query string, types []string) ([]*DNSResult, error) {
	if len(types) == 0 {
		types = []string{"A"}
	}

	var results []*DNSResult

	for _, t := range types {
		result := &DNSResult{
			QueryType: t,
			Query:     query,
		}

		switch strings.ToUpper(t) {
		case "A":
			ips, err := net.LookupIP(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				for _, ip := range ips {
					if ip4 := ip.To4(); ip4 != nil {
						result.Results = append(result.Results, ip4.String())
					}
				}
			}
		case "AAAA":
			ips, err := net.LookupIP(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				for _, ip := range ips {
					if ip.To4() == nil {
						result.Results = append(result.Results, ip.String())
					}
				}
			}
		case "MX":
			mxs, err := net.LookupMX(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				for _, mx := range mxs {
					result.Results = append(result.Results, fmt.Sprintf("%s (pref=%d)", mx.Host, mx.Pref))
				}
			}
		case "NS":
			nss, err := net.LookupNS(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				for _, ns := range nss {
					result.Results = append(result.Results, ns.Host)
				}
			}
		case "TXT":
			txts, err := net.LookupTXT(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Results = txts
			}
		case "PTR":
			names, err := net.LookupAddr(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Results = names
			}
		case "CNAME":
			cname, err := net.LookupCNAME(query)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Results = append(result.Results, cname)
			}
		default:
			result.Error = fmt.Sprintf("unsupported query type: %s", t)
		}

		results = append(results, result)
	}

	return results, nil
}
