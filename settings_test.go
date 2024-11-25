package settings_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/tomr-ninja/go-settings"
)

func testParserType[T comparable](t *testing.T, expected T, builder func(setting *settings.Setting)) {
	var v T

	builder(settings.Add(&v))
	settings.MustParse()

	if v != expected {
		t.Errorf("expected %v, got %v", expected, v)
	}
}

func TestDefaultParser(t *testing.T) {
	setup := func(t *testing.T, val string) {
		t.Helper()

		t.Setenv("OPTION1", val)
		settings.DefaultParser.SetYAML(fmt.Sprintf("option1: %s", val))
		settings.DefaultParser.SetArgs([]string{fmt.Sprintf("--option1=%s", val)})
	}

	t.Run("parse", func(t *testing.T) {
		type testRun struct {
			name    string
			builder func(setting *settings.Setting)
		}

		runs := []testRun{
			{name: "yaml only", builder: func(setting *settings.Setting) { setting.YAML("option1") }},
			{name: "env only", builder: func(setting *settings.Setting) { setting.Env("OPTION1") }},
			{name: "flag only", builder: func(setting *settings.Setting) { setting.Flag("option1") }},
			{name: "all combined", builder: func(setting *settings.Setting) { setting.YAML("option1").Env("OPTION1").Flag("option1") }},
		}

		for _, run := range runs {
			t.Run("string", func(t *testing.T) {
				setup(t, "test")
				testParserType(t, "test", run.builder)
			})

			t.Run("int", func(t *testing.T) {
				setup(t, "42")
				testParserType(t, 42, run.builder)
			})

			t.Run("bool", func(t *testing.T) {
				setup(t, "true")
				testParserType(t, true, run.builder)
			})

			t.Run("float", func(t *testing.T) {
				setup(t, "3.14")
				testParserType(t, 3.14, run.builder)
			})

			t.Run("time.Duration", func(t *testing.T) {
				setup(t, "1h")
				testParserType(t, time.Hour, run.builder)
			})
		}
	})

	t.Run("required", func(t *testing.T) {
		v := ""

		settings.Add(&v).Env("MISSING_OPTION").Required(true)

		if err := settings.Parse(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("default value", func(t *testing.T) {
		v := 0

		settings.Add(&v).Env("MISSING_OPTION").Default(42)
		settings.MustParse()

		if v != 42 {
			t.Errorf("expected %d, got %d", 42, v)
		}
	})

	t.Run("big config", func(t *testing.T) {
		cfg := struct {
			Option1 string
			Option2 int
		}{}

		t.Setenv("OPTION1", "test")
		t.Setenv("OPTION2", "42")

		settings.DefaultParser.SetYAML("option1: test\noption2: 42")

		settings.Add(&cfg.Option1).YAML("option1")
		settings.Add(&cfg.Option2).Env("OPTION2")

		settings.MustParse()

		if cfg.Option1 != "test" {
			t.Errorf("expected %s, got %s", "test", cfg.Option1)
		}
		if cfg.Option2 != 42 {
			t.Errorf("expected %d, got %d", 42, cfg.Option2)
		}
	})
}

func TestCustomParser(t *testing.T) {
	t.Setenv("OPTION1", "test")

	v := ""
	parser := settings.NewParser()
	parser.Add(&v).Env("OPTION1")

	if err := parser.Parse(); err != nil {
		t.Error(err)
	}
	if v != "test" {
		t.Errorf("expected %s, got %s", "test", v)
	}
}
