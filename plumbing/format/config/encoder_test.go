package config

import (
	"bytes"
	"fmt"
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

func (s *EncoderSuite) TestEncodeRefusesNamesGitRefuses() {
	tests := []struct {
		name string
		cfg  *Config
		want error
	}{
		{
			name: "key with newline would inject an option",
			cfg:  New().AddOption("core", "", "safe", "1").AddOption("core", "", "x\n\tinjected", "pwned"),
			want: ErrInvalidKey,
		},
		{
			name: "key with newline in subsection",
			cfg:  New().AddOption("remote", "origin", "url\nevil", "x"),
			want: ErrInvalidKey,
		},
		{name: "empty key", cfg: New().AddOption("core", "", "", "v"), want: ErrInvalidKey},
		{name: "key starting with digit", cfg: New().AddOption("core", "", "1foo", "v"), want: ErrInvalidKey},
		{name: "key with underscore", cfg: New().AddOption("core", "", "foo_bar", "v"), want: ErrInvalidKey},
		{name: "key with space", cfg: New().AddOption("core", "", "foo bar", "v"), want: ErrInvalidKey},
		{name: "key with dot", cfg: New().AddOption("core", "", "foo.bar", "v"), want: ErrInvalidKey},
		{name: "key with equals", cfg: New().AddOption("core", "", "foo=bar", "v"), want: ErrInvalidKey},
		{name: "key with non-ASCII letter", cfg: New().AddOption("core", "", "föo", "v"), want: ErrInvalidKey},
		{
			name: "subsection with newline",
			cfg:  New().AddOption("remote", "a\nb", "url", "x"),
			want: ErrInvalidSubsectionName,
		},
		{
			name: "subsection with newline and no options",
			cfg:  configWithEmptySubsection("remote", "a\n[core]"),
			want: ErrInvalidSubsectionName,
		},
		{name: "empty section", cfg: New().AddOption("", "", "k", "v"), want: ErrInvalidSectionName},
		{name: "section with newline", cfg: New().AddOption("core]\n[evil", "", "k", "v"), want: ErrInvalidSectionName},
		{name: "section with underscore", cfg: New().AddOption("co_re", "", "k", "v"), want: ErrInvalidSectionName},
		{name: "section with space", cfg: New().AddOption("co re", "", "k", "v"), want: ErrInvalidSectionName},
		{name: "section with closing bracket", cfg: New().AddOption("core]", "", "k", "v"), want: ErrInvalidSectionName},
	}

	for _, tc := range tests {
		buf := &bytes.Buffer{}
		err := NewEncoder(buf).Encode(tc.cfg)
		s.ErrorIs(err, tc.want, tc.name)
		s.Empty(buf.String(), "%s: nothing may be written for a refused config", tc.name)
	}
}

func (s *EncoderSuite) TestEncodeAcceptsNamesGitAccepts() {
	cfg := New().
		AddOption("core-1", "", "foo-bar", "1").
		AddOption("Core-X", "", "Key9", "2").
		AddOption("remote", `we"ird\name]`, "url", "3").
		AddOption("remote", "tab\tand space", "url", "4").
		AddOption("remote", "", "url", "5")

	buf := &bytes.Buffer{}
	s.Require().NoError(NewEncoder(buf).Encode(cfg))

	decoded := New()
	s.Require().NoError(NewDecoder(buf).Decode(decoded))
	s.Equal("3", decoded.Section("remote").Subsection(`we"ird\name]`).Option("url"))
	s.Equal("4", decoded.Section("remote").Subsection("tab\tand space").Option("url"))
	s.Equal("1", decoded.Section("core-1").Option("foo-bar"))
}

func configWithEmptySubsection(section, subsection string) *Config {
	cfg := New()
	cfg.Section(section).Subsection(subsection)
	return cfg
}
