package config

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EncoderSuite struct {
	suite.Suite
}

func TestEncoderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EncoderSuite))
}

func (s *EncoderSuite) TestEncode() {
	for idx, fixture := range fixtures {
		buf := &bytes.Buffer{}
		e := NewEncoder(buf)
		err := e.Encode(fixture.Config)
		s.NoError(err, fmt.Sprintf("encoder error for fixture: %d", idx))
		s.Equal(fixture.Text, buf.String(), fmt.Sprintf("bad result for fixture: %d", idx))
	}
}

func (s *EncoderSuite) TestEncodeQuotesValuesGitQuotes() {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"carriage return at the end", "foo\r", "\tkey = \"foo\r\"\n"},
		{"carriage return in the middle", "foo\rbar", "\tkey = \"foo\rbar\"\n"},
		{"carriage return at the start", "\rfoo", "\tkey = \"\rfoo\"\n"},
		{"trailing space", "foo ", "\tkey = \"foo \"\n"},
		{"leading space", " foo", "\tkey = \" foo\"\n"},
		{"trailing tab", "foo\t", "\tkey = \"foo\\t\"\n"},
		{"form feed is not whitespace to git", "foo\f", "\tkey = foo\f\n"},
		{"plain value", "foo", "\tkey = foo\n"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cfg := New().AddOption("section", "", "key", tc.value)
			buf := &bytes.Buffer{}
			s.Require().NoError(NewEncoder(buf).Encode(cfg))
			s.Equal("[section]\n"+tc.want, buf.String())
		})
	}
}

func (s *EncoderSuite) TestEncodeRoundTripsValues() {
	values := []string{
		"foo ", " foo", "foo\t", "foo\f", "foo\v", "a\\b", "a\"b",
		"a;b", "a#b", "foo\n",
	}
	valuesWithCR := []string{
		"foo\r", "foo\rbar", "\rfoo", "foo\r\n", "foo\\\r", "\r", "foo\r\r",
	}

	roundTrip := func(value string) {
		buf := &bytes.Buffer{}
		s.Require().NoError(NewEncoder(buf).Encode(New().AddOption("section", "", "key", value)))

		got := New()
		s.Require().NoError(NewDecoder(buf).Decode(got))
		s.Equal(value, got.Section("section").Option("key"))
	}

	for _, value := range values {
		s.Run(fmt.Sprintf("%q", value), func() { roundTrip(value) })
	}

	for _, value := range valuesWithCR {
		s.Run(fmt.Sprintf("%q", value), func() {
			if !readerKeepsQuotedCR() {
				s.T().Skip("gcfg drops a CR inside a quoted value; fixed by go-git/gcfg#20")
			}
			roundTrip(value)
		})
	}
}

func readerKeepsQuotedCR() bool {
	cfg := New()
	err := NewDecoder(strings.NewReader("[section]\n\tkey = \"a\rb\"\n")).Decode(cfg)
	return err == nil && cfg.Section("section").Option("key") == "a\rb"
}
