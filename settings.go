package settings

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/goccy/go-yaml"
)

var DefaultParser = NewParser()

var (
	ErrUnsupportedType       = errors.New("settings: unsupported type")
	ErrNilPointer            = errors.New("settings: nil pointer")
	ErrRequiredFieldNotFound = errors.New("settings: required field not found")
	ErrNotFound              = errors.New("settings: setting not found")
)

func init() {
	if len(os.Args) > 1 {
		DefaultParser.SetArgs(os.Args[1:])
	}
}

type Parser struct {
	yaml      string
	envPrefix string
	args      []string

	settings []*Setting
}

type Option func(p *Parser)

func NewParser(opts ...Option) *Parser {
	p := new(Parser)
	for _, o := range opts {
		o(p)
	}

	return p
}

func WithYAML(yaml string) Option {
	return func(p *Parser) {
		p.yaml = yaml
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(p *Parser) {
		p.envPrefix = prefix
	}
}

func WithArgs(args []string) Option {
	return func(p *Parser) {
		p.args = args
	}
}

func (p *Parser) SetYAML(yaml string) {
	p.yaml = yaml
}

func (p *Parser) ReadYAMLFile(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	p.yaml = string(f)

	return nil
}

func (p *Parser) SetEnvPrefix(prefix string) {
	p.envPrefix = prefix
}

func (p *Parser) SetArgs(args []string) {
	p.args = args
}

func (p *Parser) Add(v any) *Setting {
	s := new(Setting)
	s.v = v
	p.settings = append(p.settings, s)

	return s
}

func (p *Parser) Reset() {
	p.yaml = ""
	p.envPrefix = ""
	p.args = nil
	p.settings = nil
}

func (p *Parser) Parse() error {
	for _, s := range p.settings {
		if err := p.set(s); err != nil {
			p.Reset()

			return err
		}
	}

	p.Reset()

	return nil
}

func (p *Parser) MustParse() {
	if err := p.Parse(); err != nil {
		panic(err)
	}
}

func (p *Parser) set(s *Setting) error {
	if s.v == nil {
		return ErrNilPointer
	}

	parsed, err := p.parse(s)
	if err != nil {
		return err
	}

	if !parsed && s.required {
		return ErrRequiredFieldNotFound
	}

	if !parsed && s.defaultValue != nil {
		switch v := s.v.(type) {
		case *string:
			*v = s.defaultValue.(string)
		case *int:
			*v = s.defaultValue.(int)
		case *float64:
			*v = s.defaultValue.(float64)
		case *bool:
			*v = s.defaultValue.(bool)
		case *time.Duration:
			*v = s.defaultValue.(time.Duration)
		default:
			return ErrUnsupportedType
		}
	}

	return nil
}

func (p *Parser) parse(s *Setting) (parsed bool, err error) {
	for _, f := range s.parseFuncs {
		if err = f(p); err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}

			return false, err
		}

		return true, nil
	}

	return false, nil
}

func parseYAML(v any, yamlPath string, yamlData string) error {
	if len(yamlData) == 0 {
		return ErrNotFound
	}

	path, err := yaml.PathString(yamlPath)
	if err != nil {
		return err
	}

	if err = path.Read(bytes.NewBufferString(yamlData), v); err != nil {
		return err
	}

	return nil
}

func parseEnv(v any, envVar string) error {
	vStr, ok := os.LookupEnv(envVar)
	if !ok {
		return ErrNotFound
	}

	if err := setFromString(v, vStr); err != nil {
		return err
	}

	return nil
}

func parseFlag(v any, f string, args []string) error {
	vStr := ""
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&vStr, f, "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if vStr == "" {
		return ErrNotFound
	}

	if err := setFromString(v, vStr); err != nil {
		return err
	}

	return nil
}

func setFromString(v any, vStr string) error {
	switch v := v.(type) {
	case *string:
		*v = vStr
	case *int:
		vInt, err := strconv.Atoi(vStr)
		if err != nil {
			return err
		}

		*v = vInt
	case *float64:
		vFloat, err := strconv.ParseFloat(vStr, 64)
		if err != nil {
			return err
		}

		*v = vFloat
	case *bool:
		vBool, err := strconv.ParseBool(vStr)
		if err != nil {
			return err
		}

		*v = vBool
	case *time.Duration:
		vDuration, err := time.ParseDuration(vStr)
		if err != nil {
			return err
		}

		*v = vDuration
	default:
		return ErrUnsupportedType
	}

	return nil
}

func Add(v any) *Setting {
	return DefaultParser.Add(v)
}

func Parse() error {
	return DefaultParser.Parse()
}

func MustParse() {
	DefaultParser.MustParse()
}

type Setting struct {
	v          any
	parseFuncs []func(*Parser) error

	defaultValue any
	required     bool
}

type FieldParseOption func(v *Setting)

func (s *Setting) Env(envVar string) *Setting {
	s.parseFuncs = append(s.parseFuncs, func(p *Parser) error {
		return parseEnv(s.v, p.envPrefix+envVar)
	})

	return s
}

func (s *Setting) YAML(yamlPath string) *Setting {
	if yamlPath[0] != '$' {
		yamlPath = "$." + yamlPath
	}

	s.parseFuncs = append(s.parseFuncs, func(p *Parser) error {
		return parseYAML(s.v, yamlPath, p.yaml)
	})

	return s
}

func (s *Setting) Flag(flag string) *Setting {
	s.parseFuncs = append(s.parseFuncs, func(p *Parser) error {
		return parseFlag(s.v, flag, p.args)
	})

	return s
}

func (s *Setting) Required(isRequired bool) *Setting {
	s.required = isRequired

	return s
}

func (s *Setting) Default(v any) *Setting {
	s.defaultValue = v

	return s
}
