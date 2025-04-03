/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package mock provides all methods required to simulate a HCloud provider environment
package mock

import (
	"reflect"
)

// manipulateStruct changes the given member of a struct for testing purposes.
//
// PARAMETERS
// structData interface{} A pointer to a struct
// key        string      Struct member name to manipulate
// value      interface{} Value to set
func manipulateStruct(structData interface{}, key string, value interface{}) {
	fieldValue := reflect.Indirect(reflect.ValueOf(structData)).FieldByName(key)

	if fieldValue.IsValid() && fieldValue.CanSet() {
		fieldValue.Set(reflect.ValueOf(value))
	}
}
