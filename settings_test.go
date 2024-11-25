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
		testParserType(t, "test", func(setting *settings.Setting) { setting.Env("MISSING_OPTION").Default("test") })
		testParserType(t, 42, func(setting *settings.Setting) { setting.Env("MISSING_OPTION").Default(42) })
		testParserType(t, true, func(setting *settings.Setting) { setting.Env("MISSING_OPTION").Default(true) })
		testParserType(t, 3.14, func(setting *settings.Setting) { setting.Env("MISSING_OPTION").Default(3.14) })
		testParserType(t, time.Hour, func(setting *settings.Setting) { setting.Env("MISSING_OPTION").Default(time.Hour) })

		t.Run("default only", func(t *testing.T) {
			testParserType(t, "test", func(setting *settings.Setting) { setting.Default("test") })
			testParserType(t, 42, func(setting *settings.Setting) { setting.Default(42) })
			testParserType(t, true, func(setting *settings.Setting) { setting.Default(true) })
			testParserType(t, 3.14, func(setting *settings.Setting) { setting.Default(3.14) })
			testParserType(t, time.Hour, func(setting *settings.Setting) { setting.Default(time.Hour) })
		})
	})

	t.Run("big config", func(t *testing.T) {
		cfg := struct {
			Option1 string
			Option2 int
			Option3 bool
			Option4 float64
			Option5 time.Duration
		}{}

		t.Setenv("OPTION1", "test1")
		t.Setenv("OPTION2", "42")
		t.Setenv("OPTION3", "false")
		t.Setenv("OPTION4", "3.14")
		t.Setenv("OPTION5", "1h")

		settings.DefaultParser.SetYAML("option1: test2\noption2: 43\noption3: true\noption4: 3.15\noption5: 1h1m")

		settings.Add(&cfg.Option1).YAML("option1").Env("OPTION1")
		settings.Add(&cfg.Option2).Env("OPTION2").YAML("option2")
		settings.Add(&cfg.Option3).Env("OPTION3").YAML("option3")
		settings.Add(&cfg.Option4).YAML("option4").Env("OPTION4")
		settings.Add(&cfg.Option5).Env("OPTION5").YAML("option5")

		settings.MustParse()

		if cfg.Option1 != "test2" {
			t.Errorf("expected %s, got %s", "test2", cfg.Option1)
		}
		if cfg.Option2 != 42 {
			t.Errorf("expected %d, got %d", 42, cfg.Option2)
		}
		if cfg.Option3 != false {
			t.Errorf("expected %v, got %v", false, cfg.Option3)
		}
		if cfg.Option4 != 3.15 {
			t.Errorf("expected %f, got %f", 3.15, cfg.Option4)
		}
		if cfg.Option5 != time.Hour {
			t.Errorf("expected %v, got %v", time.Hour, cfg.Option5)
		}
	})
}

func TestCustomParser(t *testing.T) {
	t.Setenv("TEST_OPTION1", "test_env")

	v := ""
	parser := settings.NewParser(
		settings.WithEnvPrefix("TEST_"),
		settings.WithArgs([]string{"--option1=test_flag"}),
		settings.WithYAML("option1: test_yaml"),
	)
	parser.Add(&v).Env("OPTION1").Flag("option1")

	if err := parser.Parse(); err != nil {
		t.Error(err)
	}
	if v != "test_env" {
		t.Errorf("expected %s, got %s", "test_env", v)
	}
}
