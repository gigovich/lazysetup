package lazysetup

import (
	"fmt"
	"testing"
)

func TestSettingsInitWithError(t *testing.T) {
	settings := New()

	settings.OnInit(func() error {
		return fmt.Errorf("test error")
	}, "errorFunction")

	if err := settings.Init(); err == nil {
		t.Error("erro should happen")
		return
	}
}

func TestSettingsInitSuccess(t *testing.T) {
	settings := New()

	globalVariable := "disabled"
	settings.OnInit(func() error {
		globalVariable = "enabled"
		return nil
	}, "changeGlobalVar")

	if err := settings.Init(); err != nil {
		t.Error(err)
		return
	}

	if globalVariable != "enabled" {
		t.Errorf("item value expected 'enabled', got '%v'", globalVariable)
	}
	t.Run("twice call", func(t *testing.T) {
		// call second time, also should return nil instead error
		if err := settings.Init(); err != nil {
			t.Errorf("second call should not issued error: %v", err)
		}
	})
}

func TestSettingsDependency(t *testing.T) {
	settings := New()

	var path string
	settings.OnInit(func() error {
		path += "5"
		return nil
	}, "5", "4")

	settings.OnInit(func() error {
		path += "1,"
		return nil
	}, "1")

	settings.OnInit(func() error {
		path += "4,"
		return nil
	}, "4", "3")

	settings.OnInit(func() error {
		path += "2,"
		return nil
	}, "2", "1")

	settings.OnInit(func() error {
		path += "3,"
		return nil
	}, "3", "2")

	if err := settings.Init(); err != nil {
		t.Error(err)
		return
	}

	if path != "1,2,3,4,5" {
		t.Errorf("expected call chain '1,2,3,4,5': goted '%v'", path)
	}
}
