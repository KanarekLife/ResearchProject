package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
)

// envPrefix starts a variable that overrides one config key, named after its
// path: MCBENCH_RUN_OUT is run.out and MCBENCH_PROVIDERS_PROXY_API_KEY_ENV is
// providers.proxy.api_key_env. A value other than a string is YAML, so a list
// is [a, b] and a map {X-Run: "{run}"}. The prefix is reserved: a variable that
// names no key is an error, so a typo does not go unnoticed.
const envPrefix = "MCBENCH_"

var errNoKey = errors.New("names no config key")

func applyEnv(cfg *config, environ []string) error {
	for _, kv := range environ {
		name, value, _ := strings.Cut(kv, "=")
		path, ok := strings.CutPrefix(name, envPrefix)
		if !ok {
			continue
		}
		if err := setPath(reflect.ValueOf(cfg).Elem(), strings.Split(strings.ToLower(path), "_"), value); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

// setPath sets the key that tokens name under v. YAML names may contain
// underscores themselves (api_key_env, a provider name), so each way to split
// the tokens is tried until the rest names a key.
func setPath(v reflect.Value, tokens []string, value string) error {
	switch {
	case len(tokens) == 0 && v.Kind() == reflect.String:
		v.SetString(value)
		return nil
	case len(tokens) == 0 && v.Kind() == reflect.Pointer:
		p := reflect.New(v.Type().Elem())
		if err := setPath(p.Elem(), nil, value); err != nil {
			return err
		}
		v.Set(p)
		return nil
	case len(tokens) == 0:
		p := reflect.New(v.Type())
		if err := yaml.UnmarshalWithOptions([]byte(value), p.Interface(), yaml.Strict()); err != nil {
			return err
		}
		v.Set(p.Elem())
		return nil
	case v.Kind() == reflect.Struct:
		for i := range v.NumField() {
			tag, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("yaml"), ",")
			n := strings.Count(tag, "_") + 1
			if tag == "" || tag == "-" || n > len(tokens) || strings.Join(tokens[:n], "_") != tag {
				continue
			}
			if err := setPath(v.Field(i), tokens[n:], value); !errors.Is(err, errNoKey) {
				return err
			}
		}
	case v.Kind() == reflect.Map && v.Type().Elem().Kind() == reflect.Struct:
		for n := 1; n < len(tokens); n++ {
			key := reflect.ValueOf(strings.Join(tokens[:n], "_"))
			elem := reflect.New(v.Type().Elem()).Elem()
			if old := v.MapIndex(key); old.IsValid() {
				elem.Set(old)
			}
			err := setPath(elem, tokens[n:], value)
			if errors.Is(err, errNoKey) {
				continue
			}
			if err == nil {
				if v.IsNil() {
					v.Set(reflect.MakeMap(v.Type()))
				}
				v.SetMapIndex(key, elem)
			}
			return err
		}
	}
	return errNoKey
}
