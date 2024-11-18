package settings_test

import (
	"os"
	"testing"

	"github.com/tomr-ninja/go-settings"
)

func TestMustParse(t *testing.T) {
	t.Run("parse string", func(t *testing.T) {
		_ = os.Setenv("OPTION1", "test")
		settings.DefaultParser.SetYAML("option1: test")
		settings.DefaultParser.SetArgs([]string{"--option1=test"})

		t.Run("yaml only", func(t *testing.T) {
			v := ""
			settings.MustParse(&v, settings.YAML("option1"))

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("env only", func(t *testing.T) {
			v := ""
			settings.MustParse(&v, settings.Env("OPTION1"))

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("flag only", func(t *testing.T) {
			v := ""
			settings.MustParse(&v, settings.Flag("option1"))

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("all combined", func(t *testing.T) {
			v := ""
			settings.MustParse(&v, settings.YAML("option1"), settings.Env("OPTION1"), settings.Flag("option1"))

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})
	})

	t.Run("required", func(t *testing.T) {
		v := ""
		_, err := settings.Parse(&v, settings.Env("MISSING_OPTION"), settings.Required(true))
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("default value", func(t *testing.T) {
		v := 0
		settings.MustParse(&v, settings.Env("MISSING_OPTION"), settings.Default(42))

		if v != 42 {
			t.Errorf("expected %d, got %d", 42, v)
		}
	})

	t.Run("big config", func(t *testing.T) {
		cfg := struct {
			Option1 string
			Option2 int
		}{}

		_ = os.Setenv("OPTION1", "test")
		_ = os.Setenv("OPTION2", "42")

		settings.DefaultParser.SetYAML("option1: test\noption2: 42")

		settings.MustParse(&cfg.Option1, settings.YAML("option1"))
		settings.MustParse(&cfg.Option2, settings.Env("OPTION2"))

		if cfg.Option1 != "test" {
			t.Errorf("expected %s, got %s", "test", cfg.Option1)
		}
		if cfg.Option2 != 42 {
			t.Errorf("expected %d, got %d", 42, cfg.Option2)
		}
	})
}
