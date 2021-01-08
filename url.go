package loki

import (
	"fmt"
	"net/url"
	"strings"
)

type baseURL struct {
	*url.URL
	*url.Values

}

func NewURLBuilder() *baseURL {
	return &baseURL{
		URL:    &url.URL{},
		Values: &url.Values{},
	}
}

func (baseURL *baseURL) BaseUrl(protoc ,ip string, port int) (*baseURL ,error){
	switch protoc {
	case "http","https":
	default:
		return nil , fmt.Errorf("unkown proto , select between http and https. ")
	}
	baseURL.Scheme = protoc
	baseURL.Host = fmt.Sprintf("%v:%v", ip, port)
	return baseURL , nil
}


type Query struct {
	*baseURL
	App string
	Namespace string
}

type Opt string

//exactly equal.
const Equal Opt = "="
//not equal.
const NotEqual Opt = "!="
//regex matches.
const RegexMatch Opt = "=~"
//regex does not match.
const RegexNotMatch Opt = "!~"

type Selector struct {
	Container	string
	Namespace	string
}

//Range you should use it after Query.
func (baseURL *Query) Range(r *Range) *Query {
	if r != nil {
		if r.Enable {
			baseURL.Path = "/loki/api/v1/query_range"
			if r.Start.String() != "" {
				baseURL.Set("start", fmt.Sprint(r.Start.Unix()))
			}
			if r.End.String() != "" {
				baseURL.Set("end", fmt.Sprint(r.End.Unix()))
			}
		}
	}

	return baseURL
}

//Query
func (baseURL *Query) Query( filter string) *Query {
	baseURL.Path = "/loki/api/v1/query"
	baseURL.Set("query", filter)
	return baseURL
}

func (baseURL *Query) Limit(limit int) *Query {
	baseURL.Set("limit", fmt.Sprint(limit))
	return baseURL
}

func (baseURL *Query) Direction(d Direction) *Query {
	baseURL.Set("Direction", string(d))
	return baseURL
}

func (baseURL *Query) String() string {
	var buf = &strings.Builder{}
	buf.WriteString(baseURL.URL.String())
	if baseURL.Values.Encode() != "" {
		buf.WriteString("?")
		buf.WriteString(baseURL.Values.Encode())
	}
	return buf.String()
}
