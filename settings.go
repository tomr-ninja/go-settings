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
}

type Option func(p *Parser)

func NewParser(opts ...Option) *Parser {
	p := new(Parser)
	for _, o := range opts {
		o(p)
	}

	return p
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

func (p *Parser) MustParse(v any, fs ...FieldParseOption) bool {
	parsed, err := p.Parse(v, fs...)
	if err != nil {
		panic(err)
	}

	return parsed
}

func (p *Parser) Parse(v any, fs ...FieldParseOption) (bool, error) {
	if v == nil {
		return false, ErrNilPointer
	}

	opts := new(fieldParseOptions)
	for _, f := range fs {
		f(opts)
	}

	parsed, err := p.parse(v, opts)
	if err != nil {
		return false, err
	}

	if !parsed && opts.required {
		return false, ErrRequiredFieldNotFound
	}

	if !parsed && opts.defaultValue != nil {
		switch v := v.(type) {
		case *string:
			*v = opts.defaultValue.(string)
		case *int:
			*v = opts.defaultValue.(int)
		case *float64:
			*v = opts.defaultValue.(float64)
		case *bool:
			*v = opts.defaultValue.(bool)
		case *time.Duration:
			*v = opts.defaultValue.(time.Duration)
		default:
			return false, ErrUnsupportedType
		}

		return true, nil
	}

	return parsed, nil
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

func (p *Parser) parse(v any, opts *fieldParseOptions) (parsed bool, err error) {
	if opts.yamlPath != "" {
		parsed, err = p.parseYAML(v, opts.yamlPath)
		if err != nil {
			return false, err
		}
	}

	if opts.envVar != "" {
		parsed, err = p.parseEnv(v, opts.envVar)
		if err != nil {
			return false, err
		}
	}

	if opts.flag != "" {
		parsed, err = p.parseFlag(v, opts.flag)
		if err != nil {
			return false, err
		}
	}

	return parsed, nil
}

func (p *Parser) parseYAML(v any, yamlPath string) (bool, error) {
	if len(p.yaml) == 0 {
		return false, nil
	}

	path, err := yaml.PathString(yamlPath)
	if err != nil {
		return false, err
	}

	if err = path.Read(bytes.NewBufferString(p.yaml), v); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Parser) parseEnv(v any, envVar string) (bool, error) {
	vStr, ok := os.LookupEnv(p.envPrefix + envVar)
	if !ok {
		return false, nil
	}

	if err := p.parseString(v, vStr); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Parser) parseFlag(v any, f string) (bool, error) {
	vStr := ""
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&vStr, f, "", "")
	if err := fs.Parse(p.args); err != nil {
		return false, err
	}
	if vStr == "" {
		return false, nil
	}

	if err := p.parseString(v, vStr); err != nil {
		return false, err
	}

	return true, nil
}

func (*Parser) parseString(v any, vStr string) error {
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

func Parse(v any, fs ...FieldParseOption) (bool, error) {
	return DefaultParser.Parse(v, fs...)
}

func MustParse(v any, fs ...FieldParseOption) {
	DefaultParser.MustParse(v, fs...)
}

type fieldParseOptions struct {
	envVar   string
	yamlPath string
	flag     string

	defaultValue any
	required     bool
}

type FieldParseOption func(v *fieldParseOptions)

func Env(envVar string) FieldParseOption {
	return func(v *fieldParseOptions) {
		v.envVar = envVar
	}
}

func YAML(yamlPath string) FieldParseOption {
	return func(v *fieldParseOptions) {
		if yamlPath[0] != '$' {
			yamlPath = "$." + yamlPath
		}

		v.yamlPath = yamlPath
	}
}

func Flag(flag string) FieldParseOption {
	return func(v *fieldParseOptions) {
		v.flag = flag
	}
}

func Required(isRequired bool) FieldParseOption {
	return func(v *fieldParseOptions) {
		v.required = isRequired
	}
}

func Default(v any) FieldParseOption {
	return func(opts *fieldParseOptions) {
		opts.defaultValue = v
	}
}
