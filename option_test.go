package hx_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/izumin5210/hx"
)

func TestCombineOptions(t *testing.T) {
	opt1 := hx.OptionFunc(func(c *hx.Config) error {
		c.QueryParams.Add("foo", "1")
		return nil
	})
	opt2 := hx.OptionFunc(func(c *hx.Config) error {
		c.QueryParams.Add("foo", "2")
		return nil
	})
	optErr := hx.OptionFunc(func(c *hx.Config) error {
		return errors.New("error occurred")
	})

	t.Run("valid", func(t *testing.T) {
		cfg, err := hx.NewConfig()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = cfg.Apply(hx.CombineOptions(opt1, opt2))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if got, want := cfg.QueryParams["foo"], []string{"1", "2"}; !reflect.DeepEqual(got, want) {
			t.Errorf("query foo is %v, want %v", got, want)
		}
	})

	t.Run("error", func(t *testing.T) {
		cfg, err := hx.NewConfig()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = cfg.Apply(hx.CombineOptions(opt1, optErr))
		if err == nil {
			t.Error("returned nil, want an error")
		}
	})
}
