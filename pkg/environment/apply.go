package config

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/xiatechs/jsonata-go"
)

var placeholder = regexp.MustCompile(`\$\{\{\s*([^\}]*)\s*\}\}`)

// Apply applies the environment represented by 'environment' to the input value 'val'.
// It replaces placeholders in the input value with corresponding values from the environment.
func Apply(val any, environment *Environment) (any, error) {
	environment.mu.RLock()
	defer environment.mu.RUnlock()

	v := reflect.ValueOf(val)
	result := deepUnSpecify(v)

	if err := applyRecursive(result, environment.data); err != nil {
		return nil, err
	}

	return result.Interface(), nil
}

func applyRecursive(node reflect.Value, environment map[string]any) error {
	switch node.Kind() {
	case reflect.Map:
		for _, key := range node.MapKeys() {
			value := reflect.ValueOf(node.MapIndex(key).Interface())

			switch value.Kind() {
			case reflect.String:
				strValue := value.String()
				newValue, err := replacePlaceholders(strValue, environment)
				if err != nil {
					return err
				}

				node.SetMapIndex(key, reflect.ValueOf(newValue))
			case reflect.Map, reflect.Slice:
				if err := applyRecursive(value, environment); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < node.Len(); i++ {
			elem := reflect.ValueOf(node.Index(i).Interface())
			err := applyRecursive(elem, environment)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func replacePlaceholders(input string, environment map[string]any) (any, error) {
	matches := placeholder.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return input, nil
	}
	if len(matches) == 1 {
		return evaluateExpression(matches[0][1], environment)
	}

	return placeholder.ReplaceAllStringFunc(input, func(match string) string {
		matches := placeholder.FindStringSubmatch(match)
		if len(matches) != 2 {
			return match
		}
		expression := matches[1]

		value, err := evaluateExpression(expression, environment)
		if err != nil {
			return match
		}

		return fmt.Sprintf("%v", value)
	}), nil
}

func evaluateExpression(expression string, environment map[string]any) (any, error) {
	exp, err := jsonata.Compile(expression)
	if err != nil {
		return nil, err
	}
	exp.RegisterVars(environment)
	return exp.Eval(nil)
}

func deepUnSpecify(original reflect.Value) reflect.Value {
	switch original.Kind() {
	case reflect.Map:
		result := reflect.MakeMap(reflect.MapOf(original.Type().Key(), typeAny))
		for _, key := range original.MapKeys() {
			value := original.MapIndex(key)
			result.SetMapIndex(key, deepUnSpecify(value))
		}
		return result
	case reflect.Slice:
		result := reflect.MakeSlice(reflect.SliceOf(typeAny), original.Len(), original.Len())
		for i := 0; i < original.Len(); i++ {
			result.Index(i).Set(deepUnSpecify(original.Index(i)))
		}
		return result
	default:
		return original
	}
}
