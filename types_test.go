// This code has been cherry-picked from https://github.com/go-openapi/strfmt

package strfmt

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	p, _ = time.Parse(time.RFC3339Nano, "2011-08-18T19:03:37.000000000+01:00")

	testCases = []struct {
		in     []byte    // externally sourced data -- to be unmarshalled
		time   time.Time // its representation in time.Time
		str    string    // its marshalled representation
		utcStr string    // the marshaled representation as utc
	}{
		{[]byte("2014-12-15 08:00:00"), time.Date(2014, 12, 15, 8, 0, 0, 0, time.UTC), "2014-12-15T08:00:00.000Z", "2014-12-15T08:00:00.000Z"},
		{[]byte("2014-12-15T08:00:00"), time.Date(2014, 12, 15, 8, 0, 0, 0, time.UTC), "2014-12-15T08:00:00.000Z", "2014-12-15T08:00:00.000Z"},
		{[]byte("2014-12-15T08:00"), time.Date(2014, 12, 15, 8, 0, 0, 0, time.UTC), "2014-12-15T08:00:00.000Z", "2014-12-15T08:00:00.000Z"},
		{[]byte("2014-12-15T08:00Z"), time.Date(2014, 12, 15, 8, 0, 0, 0, time.UTC), "2014-12-15T08:00:00.000Z", "2014-12-15T08:00:00.000Z"},
		{[]byte("2018-01-28T23:54Z"), time.Date(2018, 0o1, 28, 23, 54, 0, 0, time.UTC), "2018-01-28T23:54:00.000Z", "2018-01-28T23:54:00.000Z"},
		{[]byte("2014-12-15T08:00:00.000Z"), time.Date(2014, 12, 15, 8, 0, 0, 0, time.UTC), "2014-12-15T08:00:00.000Z", "2014-12-15T08:00:00.000Z"},
		{[]byte("2011-08-18T19:03:37.123000000+01:00"), time.Date(2011, 8, 18, 19, 3, 37, 123*1e6, p.Location()), "2011-08-18T19:03:37.123+01:00", "2011-08-18T18:03:37.123Z"},
		{[]byte("2011-08-18T19:03:37.123000+0100"), time.Date(2011, 8, 18, 19, 3, 37, 123*1e6, p.Location()), "2011-08-18T19:03:37.123+01:00", "2011-08-18T18:03:37.123Z"},
		{[]byte("2011-08-18T19:03:37.123+0100"), time.Date(2011, 8, 18, 19, 3, 37, 123*1e6, p.Location()), "2011-08-18T19:03:37.123+01:00", "2011-08-18T18:03:37.123Z"},
		{[]byte("2014-12-15T19:30:20Z"), time.Date(2014, 12, 15, 19, 30, 20, 0, time.UTC), "2014-12-15T19:30:20.000Z", "2014-12-15T19:30:20.000Z"},
		{[]byte("0001-01-01T00:00:00Z"), time.Time{}.UTC(), "0001-01-01T00:00:00.000Z", "0001-01-01T00:00:00.000Z"},
		{[]byte(""), time.Unix(0, 0).UTC(), "1970-01-01T00:00:00.000Z", "1970-01-01T00:00:00.000Z"},
		{[]byte(nil), time.Unix(0, 0).UTC(), "1970-01-01T00:00:00.000Z", "1970-01-01T00:00:00.000Z"},
	}
)

func TestNewDateTime(t *testing.T) {
	assert.EqualValues(t, time.Unix(0, 0).UTC(), NewDateTime())
}

func TestParseDateTime_errorCases(t *testing.T) {
	_, err := ParseDateTime("yada")
	assert.Error(t, err)
}

func TestDateTime_UnmarshalJSON(t *testing.T) {
	for caseNum, example := range testCases {
		t.Logf("Case #%d", caseNum)
		pp := NewDateTime()
		err := pp.UnmarshalJSON(esc(example.in))
		assert.NoError(t, err)
		assert.EqualValues(t, example.time, pp)
	}

	// Check UnmarshalJSON failure with no lexed items
	pp := NewDateTime()
	err := pp.UnmarshalJSON([]byte("zorg emperor"))
	assert.Error(t, err)

	// Check lexer failure
	err = pp.UnmarshalJSON([]byte(`"zorg emperor"`))
	assert.Error(t, err)

	// Check null case
	err = pp.UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
}

func TestDateTime_MarshalJSON(t *testing.T) {
	for caseNum, example := range testCases {
		t.Logf("Case #%d", caseNum)
		dt := DateTime(example.time)
		bb, err := dt.MarshalJSON()
		assert.NoError(t, err)
		assert.EqualValues(t, esc([]byte(example.str)), bb)
	}
}

func TestDateTime_MarshalJSON_Override(t *testing.T) {
	oldNormalizeMarshal := NormalizeTimeForMarshal
	defer func() {
		NormalizeTimeForMarshal = oldNormalizeMarshal
	}()

	NormalizeTimeForMarshal = func(t time.Time) time.Time {
		return t.UTC()
	}
	for caseNum, example := range testCases {
		t.Logf("Case #%d", caseNum)
		dt := DateTime(example.time.UTC())
		bb, err := dt.MarshalJSON()
		assert.NoError(t, err)
		assert.EqualValues(t, esc([]byte(example.utcStr)), bb)
	}
}

func esc(v []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte('"')
	buf.Write(v)
	buf.WriteByte('"')
	return buf.Bytes()
}
