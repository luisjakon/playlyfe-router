package playlyfe

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"

	graphql "github.com/playlyfe/go-graphql"
)

type Router map[string]interface{}

func NewRouter() Router {
	return Router{}
}

func (r Router) Register(path string, resolver interface{}) error {

	if rs, ok := resolver.(func(params *graphql.ResolveParams) (interface{}, error)); ok {
		r[path] = rs
		return nil
	}

	// TODO: more exhaustive checks
	if reflect.TypeOf(resolver).Kind() != reflect.Func {
		return errors.New("Invalid resolver function")
	}

	r[path] = resolverFn(resolver)
	return nil
}

func (r Router) Join(resolver Router) Router {
	for k, v := range resolver {
		r[k] = v
	}
	return r
}

func resolverFn(resolver interface{}) func(params *graphql.ResolveParams) (interface{}, error) {
	argsType := reflect.TypeOf(resolver).In(1)
	resolverFn := reflect.ValueOf(resolver)

	return func(params *graphql.ResolveParams) (interface{}, error) {
		args := reflect.New(argsType).Interface()

		err := mapstructure.Decode(params.Args, &args)
		// err := jsonDecode(params.Args, &args)
		if err != nil {
			return nil, err
		}

		in := []reflect.Value{
			reflect.ValueOf(params),
			reflect.ValueOf(args).Elem(),
		}

		result := resolverFn.Call(in)

		if !result[1].IsNil() {
			return result[0].Interface(), result[1].Interface().(error)
		}
		return result[0].Interface(), nil
	}
}

func jsonDecode(data interface{}, v interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
