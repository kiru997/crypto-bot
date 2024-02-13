// Copyright 2017 Matt Ho
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package swagger

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

func inspect(t reflect.Type, jsonTag string) Property {
	p := Property{
		GoType: t,
	}
	if t == reflect.TypeOf(time.Time{}) {
		p.Type = "string"
		p.Format = "date-time"
		return p
	}

	// if t == reflect.TypeOf(commondto.EpochTime{}) {
	// 	p.Type = "integer"
	// 	p.Format = "int64"
	// 	p.Example = fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	// 	return p
	// }

	k := p.GoType.Kind()
	if k == reflect.Ptr {
		if p.GoType.Elem().Kind() != reflect.Struct {
			k = p.GoType.Elem().Kind()
		}
	}

	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		p.Type = "integer"
		p.Format = "int32"
	case reflect.Int64, reflect.Uint64:
		p.Type = "integer"
		p.Format = "int64"
	case reflect.Float32:
		p.Type = "number"
		p.Format = "float"
	case reflect.Float64:
		p.Type = "number"
		p.Format = "double"
	case reflect.String:
		p.Type = "string"
		p.Format = "string"
	case reflect.Bool:
		p.Type = "boolean"
		p.Format = "boolean"
	case reflect.Struct:
		name := makeName(p.GoType)
		p.Ref = makeRef(name)

		p.Type = "object"
		p.Format = "object"
	case reflect.Map:
		p.Type = "object"
		p.Format = "map"

	case reflect.Ptr:
		p.GoType = t.Elem()
		name := makeName(p.GoType)
		p.Ref = makeRef(name)
	case reflect.Slice:
		p.Type = "array"
		p.Items = &Items{}

		p.GoType = t.Elem() // dereference the slice
		switch p.GoType.Kind() {
		case reflect.Ptr:
			p.GoType = p.GoType.Elem()
			name := makeName(p.GoType)
			p.Items.Ref = makeRef(name)

		case reflect.Struct:
			name := makeName(p.GoType)
			p.Items.Ref = makeRef(name)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
			p.Items.Type = "integer"
			p.Items.Format = "int32"

		case reflect.Int64, reflect.Uint64:
			p.Items.Type = "integer"
			p.Items.Format = "int64"

		case reflect.Float64:
			p.Items.Type = "number"
			p.Items.Format = "double"

		case reflect.Float32:
			p.Items.Type = "number"
			p.Items.Format = "float"

		case reflect.String:
			p.Items.Type = "string"
			p.Items.Format = "string"
		}
	default:
		p.Type = "string"
		p.Format = "string"
	}

	return p
}

func defineObject(v interface{}) Object {
	var required []string

	var t reflect.Type
	switch value := v.(type) {
	case reflect.Type:
		t = value
	default:
		t = reflect.TypeOf(v)
	}

	properties := map[string]Property{}
	isArray := t.Kind() == reflect.Slice

	if isArray {
		t = t.Elem()
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		p := inspect(t, "")
		return Object{
			IsArray:  isArray,
			GoType:   t,
			Type:     p.Type,
			Format:   p.Format,
			Name:     t.Kind().String(),
			Required: required,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// skip unexported fields
		if strings.ToLower(field.Name[0:1]) == field.Name[0:1] {
			continue
		}

		// determine the json name of the field
		name := strings.TrimSpace(field.Tag.Get("json"))
		if name == "" || strings.HasPrefix(name, ",") {
			name = field.Name
		} else {
			// strip out things like , omitempty
			parts := strings.Split(name, ",")
			name = parts[0]
		}

		parts := strings.Split(name, ",") // foo,omitempty => foo
		name = parts[0]
		if name == "-" {
			// honor json ignore tag
			continue
		}

		// determine if this field is required or not
		if v := field.Tag.Get("required"); v == "true" {
			if required == nil {
				required = []string{}
			}
			required = append(required, name)
		}

		if field.Type.Kind() == reflect.Map {
			temp, ok := mapType[field.Type.Key().Name()+":"+field.Type.Elem().Name()]
			if ok {
				field.Type = reflect.TypeOf(temp)
			}
		}

		p := inspect(field.Type, field.Tag.Get("json"))
		p.Deprecated, _ = strconv.ParseBool(field.Tag.Get("deprecated"))
		properties[name] = p
	}

	return Object{
		IsArray:    isArray,
		GoType:     t,
		Type:       "object",
		Name:       makeName(t),
		Required:   required,
		Properties: properties,
	}
}

type typeMapStringInt64 struct {
	Property1 int64 `json:"property_1"`
	Property2 int64 `json:"property_2"`
	Property3 int64 `json:"property_3"`
}

var mapType = map[string]interface{}{
	"string:int":   typeMapStringInt64{},
	"string:int64": typeMapStringInt64{},
}

func define(v interface{}) map[string]Object {
	objMap := map[string]Object{}

	obj := defineObject(v)
	objMap[obj.Name] = obj

	dirty := true

	for dirty {
		dirty = false
		for _, d := range objMap {
			for _, p := range d.Properties {
				if p.GoType.Kind() == reflect.Struct {
					name := makeName(p.GoType)
					if _, exists := objMap[name]; !exists {
						child := defineObject(p.GoType)
						objMap[child.Name] = child
						dirty = true
					}
				}
			}
		}
	}

	return objMap
}

// MakeSchema takes struct or pointer to a struct and returns a Schema instance suitable for use by the swagger doc
func MakeSchema(prototype interface{}) *Schema {
	schema := &Schema{
		Prototype: prototype,
	}

	obj := defineObject(prototype)
	if obj.IsArray {
		schema.Type = "array"
		schema.Items = &Items{
			Ref: makeRef(obj.Name),
		}

	} else {
		schema.Ref = makeRef(obj.Name)
	}

	return schema
}
