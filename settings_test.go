package settings_test

import (
	"os"
	"testing"

	"github.com/tomr-ninja/go-settings"
)

func TestParser(t *testing.T) {
	t.Run("parse string", func(t *testing.T) {
		setup := func(t *testing.T) {
			t.Helper()

			t.Setenv("OPTION1", "test")
			settings.DefaultParser.SetYAML("option1: test")
			settings.DefaultParser.SetArgs([]string{"--option1=test"})
		}

		t.Run("yaml only", func(t *testing.T) {
			setup(t)

			v := ""
			settings.Add(&v).YAML("option1")
			settings.MustParse()

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("env only", func(t *testing.T) {
			setup(t)

			v := ""
			settings.Add(&v).Env("OPTION1")
			settings.MustParse()

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("flag only", func(t *testing.T) {
			setup(t)

			v := ""
			settings.Add(&v).Flag("option1")
			settings.MustParse()

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})

		t.Run("all combined", func(t *testing.T) {
			setup(t)

			v := ""
			settings.Add(&v).YAML("option1").Env("OPTION1").Flag("option1")
			settings.MustParse()

			if v != "test" {
				t.Errorf("expected %s, got %s", "test", v)
			}
		})
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

		_ = os.Setenv("OPTION1", "test")
		_ = os.Setenv("OPTION2", "42")

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
