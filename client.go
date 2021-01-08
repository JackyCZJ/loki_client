package loki

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type Log struct {
	//@inject_tag:json:"timestamp"
	Timestamp uint64 `json:"timestamp"`
	//@inject_tag:json:"content"
	Content string `json:"content"`
}

type Client interface {
	//LogsLeast get log by filter and limit
	LogsLeast(limit int, order Direction,query string) ([]Log, error)
	//LogsRange get log by LogsLeast and range time
	LogsRange(limit int, order Direction, r *Range,query string) ([]Log, error)
}

type lokiClient struct {
	ip     string
	port   int
	protoc string
	client fasthttp.Client
	*sync.Mutex
}

func NewLokiClient(ip string, port int) Client {
	return &lokiClient{
		ip:     ip,
		port:   port,
		client: fasthttp.Client{},
		Mutex:  &sync.Mutex{},
	}
}


func (l *lokiClient) url(limit int, order Direction, r *Range) *Query {
	if b , err := NewURLBuilder().BaseUrl(l.protoc,l.ip, l.port);err != nil{
		return nil
	}else{
		var q Query
		q.baseURL = b
		return q.Direction(order).Limit(limit).Range(r)
	}
}
func (l *lokiClient) LogsLeast(limit int, order Direction,query string) ([]Log, error) {
	body, err := l.doRequest(l.url(limit, order, nil), query)
	if err != nil {
		return nil, err
	}

	var lokiLog Loki
	err = json.Unmarshal(body, &lokiLog)
	if err != nil {
		return nil, err
	}
	if lokiLog.Status != "success" {
		return nil, err
	}
	var s Streams

	err = json.Unmarshal(lokiLog.Data.Result, &s)
	if err != nil {
		return nil, err
	}
	var logArray []Log
	for _, data := range s {
		for _, data := range data.Values {
			logArray = append(logArray,Log{
				Timestamp: uint64(data.Timestamp.UnixNano()),
				Content:   data.Line,
			})
		}

	}

	return logArray, nil
}

func (l *lokiClient) LogsRange(limit int, order Direction, r *Range,  query string) ([]Log, error) {
	var (
		total     = 0
		batchSize = 1000
	)
	var logArray []Log
	var lastEntry []*Entry
	if batchSize > limit {
		batchSize = limit
	}
	for total < limit {
		bs := batchSize
		if limit-total < batchSize {
			bs = limit - total + len(lastEntry)
		}
		body, err := l.doRequest(l.url(bs, order, r), query)
		if err != nil {
			return nil, err
		}
		var lokiLog Loki
		err = json.Unmarshal(body, &lokiLog)
		if err != nil {
			return nil, err
		}
		if lokiLog.Status != "success" {
			return nil, err
		}
		var s Streams
		err = json.Unmarshal(lokiLog.Data.Result, &s)
		if err != nil {
			return nil, err
		}
		resultLength, lastEntry, result := parseStream(s, order, lastEntry)
		for _, r := range result {
			logArray = append(logArray,Log{
				Timestamp: uint64(r.Timestamp.UnixNano()),
				Content:   r.Line,
			})
		}
		if resultLength <= 0 {
			break
		}

		if len(lastEntry) == 0 {
			break
		}

		if resultLength == limit {
			break
		}
		total += resultLength
		if order == FORWARD {
			r.Start = lastEntry[0].Timestamp
		} else {
			r.End = lastEntry[0].Timestamp.Add(1 * time.Nanosecond)
		}

	}

	return logArray, nil
}

func parseStream(streams Streams, order Direction, lastEntry []*Entry) (length int, lel []*Entry, result []*Entry) {
	allEntries := make([]streamEntryPair, 0)
	for _, s := range streams {
		for _, e := range s.Values {
			allEntries = append(allEntries, streamEntryPair{
				entry:  e,
				labels: s.Labels,
			})
		}
	}
	if len(allEntries) == 0 {
		return 0, nil, nil
	}
	if order == FORWARD {
		sort.Slice(allEntries, func(i, j int) bool { return allEntries[i].entry.Timestamp.Before(allEntries[j].entry.Timestamp) })
	} else {
		sort.Slice(allEntries, func(i, j int) bool { return allEntries[i].entry.Timestamp.After(allEntries[j].entry.Timestamp) })
	}
	for _, e := range allEntries {
		if len(lastEntry) > 0 && e.entry.Timestamp == lastEntry[0].Timestamp {
			skip := false
			for _, le := range lastEntry {
				if e.entry.Line == le.Line {
					skip = true
				}
			}
			if skip {
				continue
			}
		}
		result = append(result, &e.entry)
		length++
	}
	lel = []*Entry{}
	le := allEntries[len(allEntries)-1].entry
	for i, e := range allEntries {
		if e.entry.Timestamp.Equal(le.Timestamp) {
			lel = append(lel, &allEntries[i].entry)
		}
	}

	return length, lel, result
}

func (l *lokiClient) doRequest(baseURL *Query,query string) ([]byte, error) {
	l.Lock()
	defer l.Unlock()
	var req fasthttp.Request
	req.SetRequestURI(baseURL.Query(query).String())
	req.Header.SetMethod("GET")
	var resp fasthttp.Response
	err := l.client.Do(&req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Error response from log server: %s (%v) ", string(resp.Body()), err)
	}
	return resp.Body(), nil
}
