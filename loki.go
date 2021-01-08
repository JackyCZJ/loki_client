package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

type Loki struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}
//copy from loki cli
type LabelSet map[string]string

// String implements the Stringer interface.  It returns a formatted/sorted set of label key/value pairs.
func (l LabelSet) String() string {
	var b bytes.Buffer

	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(strconv.Quote(l[k]))
	}
	b.WriteByte('}')
	return b.String()
}

func (l LabelSet) Map() map[string]string {
	return l
}

type streamEntryPair struct {
	entry  Entry
	labels LabelSet
}
type Data struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}

type Stream struct {
	Values []Entry  `json:"values"`
	Labels LabelSet `json:"stream"`
}

type Entry struct {
	Timestamp time.Time
	Line      string
}

type Direction string

const FORWARD Direction = "FORWARD"
const BACKWARD Direction = "BACKWARD"

type Range struct {
	Start  time.Time
	End    time.Time
	Enable bool
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entry) MarshalJSON() ([]byte, error) {
	l, err := json.Marshal(e.Line)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("[\"%d\",%s]", e.Timestamp.UnixNano(), l)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Entry) UnmarshalJSON(data []byte) error {
	var unmarshal []string

	err := json.Unmarshal(data, &unmarshal)
	if err != nil {
		return err
	}

	t, err := strconv.ParseInt(unmarshal[0], 10, 64)
	if err != nil {
		return err
	}

	e.Timestamp = time.Unix(0, t)
	e.Line = unmarshal[1]

	return nil
}

type Streams []Stream
