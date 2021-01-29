/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause */

// Code generated. DO NOT EDIT.

/*
 * Data type definitions file for service: FloodProtectionProfiles.
 * Includes binding types of a structures and enumerations defined in the service.
 * Shared by client-side stubs and server-side skeletons to ensure type
 * compatibility.
 */

package infra

import (
	"reflect"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol"
)





func floodProtectionProfilesDeleteInputType() bindings.StructType {
	fields := make(map[string]bindings.BindingType)
	fieldNameMap := make(map[string]string)
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	var validators = []bindings.Validator{}
	return bindings.NewStructType("operation-input", fields, reflect.TypeOf(data.StructValue{}), fieldNameMap, validators)
}

func floodProtectionProfilesDeleteOutputType() bindings.BindingType {
	return bindings.NewVoidType()
}

func floodProtectionProfilesDeleteRestMetadata() protocol.OperationRestMetadata {
	fields := map[string]bindings.BindingType{}
	fieldNameMap := map[string]string{}
	paramsTypeMap := map[string]bindings.BindingType{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	headerParams := map[string]string{}
	dispatchHeaderParams := map[string]string{}
	bodyFieldsMap := map[string]string{}
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	paramsTypeMap["flood_protection_profile_id"] = bindings.NewStringType()
	paramsTypeMap["floodProtectionProfileId"] = bindings.NewStringType()
	pathParams["flood_protection_profile_id"] = "floodProtectionProfileId"
	resultHeaders := map[string]string{}
	errorHeaders := map[string]map[string]string{}
	return protocol.NewOperationRestMetadata(
		fields,
		fieldNameMap,
		paramsTypeMap,
		pathParams,
		queryParams,
		headerParams,
		dispatchHeaderParams,
		bodyFieldsMap,
		"",
		"",
		"DELETE",
		"/policy/api/v1/infra/flood-protection-profiles/{floodProtectionProfileId}",
		"",
		resultHeaders,
		204,
		"",
		errorHeaders,
		map[string]int{"com.vmware.vapi.std.errors.invalid_request": 400,"com.vmware.vapi.std.errors.unauthorized": 403,"com.vmware.vapi.std.errors.service_unavailable": 503,"com.vmware.vapi.std.errors.internal_server_error": 500,"com.vmware.vapi.std.errors.not_found": 404})
}

func floodProtectionProfilesGetInputType() bindings.StructType {
	fields := make(map[string]bindings.BindingType)
	fieldNameMap := make(map[string]string)
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	var validators = []bindings.Validator{}
	return bindings.NewStructType("operation-input", fields, reflect.TypeOf(data.StructValue{}), fieldNameMap, validators)
}

func floodProtectionProfilesGetOutputType() bindings.BindingType {
	return bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
}

func floodProtectionProfilesGetRestMetadata() protocol.OperationRestMetadata {
	fields := map[string]bindings.BindingType{}
	fieldNameMap := map[string]string{}
	paramsTypeMap := map[string]bindings.BindingType{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	headerParams := map[string]string{}
	dispatchHeaderParams := map[string]string{}
	bodyFieldsMap := map[string]string{}
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	paramsTypeMap["flood_protection_profile_id"] = bindings.NewStringType()
	paramsTypeMap["floodProtectionProfileId"] = bindings.NewStringType()
	pathParams["flood_protection_profile_id"] = "floodProtectionProfileId"
	resultHeaders := map[string]string{}
	errorHeaders := map[string]map[string]string{}
	return protocol.NewOperationRestMetadata(
		fields,
		fieldNameMap,
		paramsTypeMap,
		pathParams,
		queryParams,
		headerParams,
		dispatchHeaderParams,
		bodyFieldsMap,
		"",
		"",
		"GET",
		"/policy/api/v1/infra/flood-protection-profiles/{floodProtectionProfileId}",
		"",
		resultHeaders,
		200,
		"",
		errorHeaders,
		map[string]int{"com.vmware.vapi.std.errors.invalid_request": 400,"com.vmware.vapi.std.errors.unauthorized": 403,"com.vmware.vapi.std.errors.service_unavailable": 503,"com.vmware.vapi.std.errors.internal_server_error": 500,"com.vmware.vapi.std.errors.not_found": 404})
}

func floodProtectionProfilesListInputType() bindings.StructType {
	fields := make(map[string]bindings.BindingType)
	fieldNameMap := make(map[string]string)
	fields["cursor"] = bindings.NewOptionalType(bindings.NewStringType())
	fields["include_mark_for_delete_objects"] = bindings.NewOptionalType(bindings.NewBooleanType())
	fields["included_fields"] = bindings.NewOptionalType(bindings.NewStringType())
	fields["page_size"] = bindings.NewOptionalType(bindings.NewIntegerType())
	fields["sort_ascending"] = bindings.NewOptionalType(bindings.NewBooleanType())
	fields["sort_by"] = bindings.NewOptionalType(bindings.NewStringType())
	fieldNameMap["cursor"] = "Cursor"
	fieldNameMap["include_mark_for_delete_objects"] = "IncludeMarkForDeleteObjects"
	fieldNameMap["included_fields"] = "IncludedFields"
	fieldNameMap["page_size"] = "PageSize"
	fieldNameMap["sort_ascending"] = "SortAscending"
	fieldNameMap["sort_by"] = "SortBy"
	var validators = []bindings.Validator{}
	return bindings.NewStructType("operation-input", fields, reflect.TypeOf(data.StructValue{}), fieldNameMap, validators)
}

func floodProtectionProfilesListOutputType() bindings.BindingType {
	return bindings.NewReferenceType(model.FloodProtectionProfileListResultBindingType)
}

func floodProtectionProfilesListRestMetadata() protocol.OperationRestMetadata {
	fields := map[string]bindings.BindingType{}
	fieldNameMap := map[string]string{}
	paramsTypeMap := map[string]bindings.BindingType{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	headerParams := map[string]string{}
	dispatchHeaderParams := map[string]string{}
	bodyFieldsMap := map[string]string{}
	fields["cursor"] = bindings.NewOptionalType(bindings.NewStringType())
	fields["include_mark_for_delete_objects"] = bindings.NewOptionalType(bindings.NewBooleanType())
	fields["included_fields"] = bindings.NewOptionalType(bindings.NewStringType())
	fields["page_size"] = bindings.NewOptionalType(bindings.NewIntegerType())
	fields["sort_ascending"] = bindings.NewOptionalType(bindings.NewBooleanType())
	fields["sort_by"] = bindings.NewOptionalType(bindings.NewStringType())
	fieldNameMap["cursor"] = "Cursor"
	fieldNameMap["include_mark_for_delete_objects"] = "IncludeMarkForDeleteObjects"
	fieldNameMap["included_fields"] = "IncludedFields"
	fieldNameMap["page_size"] = "PageSize"
	fieldNameMap["sort_ascending"] = "SortAscending"
	fieldNameMap["sort_by"] = "SortBy"
	paramsTypeMap["included_fields"] = bindings.NewOptionalType(bindings.NewStringType())
	paramsTypeMap["page_size"] = bindings.NewOptionalType(bindings.NewIntegerType())
	paramsTypeMap["include_mark_for_delete_objects"] = bindings.NewOptionalType(bindings.NewBooleanType())
	paramsTypeMap["cursor"] = bindings.NewOptionalType(bindings.NewStringType())
	paramsTypeMap["sort_by"] = bindings.NewOptionalType(bindings.NewStringType())
	paramsTypeMap["sort_ascending"] = bindings.NewOptionalType(bindings.NewBooleanType())
	queryParams["cursor"] = "cursor"
	queryParams["sort_ascending"] = "sort_ascending"
	queryParams["included_fields"] = "included_fields"
	queryParams["sort_by"] = "sort_by"
	queryParams["include_mark_for_delete_objects"] = "include_mark_for_delete_objects"
	queryParams["page_size"] = "page_size"
	resultHeaders := map[string]string{}
	errorHeaders := map[string]map[string]string{}
	return protocol.NewOperationRestMetadata(
		fields,
		fieldNameMap,
		paramsTypeMap,
		pathParams,
		queryParams,
		headerParams,
		dispatchHeaderParams,
		bodyFieldsMap,
		"",
		"",
		"GET",
		"/policy/api/v1/infra/flood-protection-profiles",
		"",
		resultHeaders,
		200,
		"",
		errorHeaders,
		map[string]int{"com.vmware.vapi.std.errors.invalid_request": 400,"com.vmware.vapi.std.errors.unauthorized": 403,"com.vmware.vapi.std.errors.service_unavailable": 503,"com.vmware.vapi.std.errors.internal_server_error": 500,"com.vmware.vapi.std.errors.not_found": 404})
}

func floodProtectionProfilesPatchInputType() bindings.StructType {
	fields := make(map[string]bindings.BindingType)
	fieldNameMap := make(map[string]string)
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fields["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	fieldNameMap["flood_protection_profile"] = "FloodProtectionProfile"
	var validators = []bindings.Validator{}
	return bindings.NewStructType("operation-input", fields, reflect.TypeOf(data.StructValue{}), fieldNameMap, validators)
}

func floodProtectionProfilesPatchOutputType() bindings.BindingType {
	return bindings.NewVoidType()
}

func floodProtectionProfilesPatchRestMetadata() protocol.OperationRestMetadata {
	fields := map[string]bindings.BindingType{}
	fieldNameMap := map[string]string{}
	paramsTypeMap := map[string]bindings.BindingType{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	headerParams := map[string]string{}
	dispatchHeaderParams := map[string]string{}
	bodyFieldsMap := map[string]string{}
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fields["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	fieldNameMap["flood_protection_profile"] = "FloodProtectionProfile"
	paramsTypeMap["flood_protection_profile_id"] = bindings.NewStringType()
	paramsTypeMap["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	paramsTypeMap["floodProtectionProfileId"] = bindings.NewStringType()
	pathParams["flood_protection_profile_id"] = "floodProtectionProfileId"
	resultHeaders := map[string]string{}
	errorHeaders := map[string]map[string]string{}
	return protocol.NewOperationRestMetadata(
		fields,
		fieldNameMap,
		paramsTypeMap,
		pathParams,
		queryParams,
		headerParams,
		dispatchHeaderParams,
		bodyFieldsMap,
		"",
		"flood_protection_profile",
		"PATCH",
		"/policy/api/v1/infra/flood-protection-profiles/{floodProtectionProfileId}",
		"",
		resultHeaders,
		204,
		"",
		errorHeaders,
		map[string]int{"com.vmware.vapi.std.errors.invalid_request": 400,"com.vmware.vapi.std.errors.unauthorized": 403,"com.vmware.vapi.std.errors.service_unavailable": 503,"com.vmware.vapi.std.errors.internal_server_error": 500,"com.vmware.vapi.std.errors.not_found": 404})
}

func floodProtectionProfilesUpdateInputType() bindings.StructType {
	fields := make(map[string]bindings.BindingType)
	fieldNameMap := make(map[string]string)
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fields["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	fieldNameMap["flood_protection_profile"] = "FloodProtectionProfile"
	var validators = []bindings.Validator{}
	return bindings.NewStructType("operation-input", fields, reflect.TypeOf(data.StructValue{}), fieldNameMap, validators)
}

func floodProtectionProfilesUpdateOutputType() bindings.BindingType {
	return bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
}

func floodProtectionProfilesUpdateRestMetadata() protocol.OperationRestMetadata {
	fields := map[string]bindings.BindingType{}
	fieldNameMap := map[string]string{}
	paramsTypeMap := map[string]bindings.BindingType{}
	pathParams := map[string]string{}
	queryParams := map[string]string{}
	headerParams := map[string]string{}
	dispatchHeaderParams := map[string]string{}
	bodyFieldsMap := map[string]string{}
	fields["flood_protection_profile_id"] = bindings.NewStringType()
	fields["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	fieldNameMap["flood_protection_profile_id"] = "FloodProtectionProfileId"
	fieldNameMap["flood_protection_profile"] = "FloodProtectionProfile"
	paramsTypeMap["flood_protection_profile_id"] = bindings.NewStringType()
	paramsTypeMap["flood_protection_profile"] = bindings.NewDynamicStructType([]bindings.ReferenceType{bindings.NewReferenceType(model.FloodProtectionProfileBindingType),}, bindings.REST)
	paramsTypeMap["floodProtectionProfileId"] = bindings.NewStringType()
	pathParams["flood_protection_profile_id"] = "floodProtectionProfileId"
	resultHeaders := map[string]string{}
	errorHeaders := map[string]map[string]string{}
	return protocol.NewOperationRestMetadata(
		fields,
		fieldNameMap,
		paramsTypeMap,
		pathParams,
		queryParams,
		headerParams,
		dispatchHeaderParams,
		bodyFieldsMap,
		"",
		"flood_protection_profile",
		"PUT",
		"/policy/api/v1/infra/flood-protection-profiles/{floodProtectionProfileId}",
		"",
		resultHeaders,
		200,
		"",
		errorHeaders,
		map[string]int{"com.vmware.vapi.std.errors.invalid_request": 400,"com.vmware.vapi.std.errors.unauthorized": 403,"com.vmware.vapi.std.errors.service_unavailable": 503,"com.vmware.vapi.std.errors.internal_server_error": 500,"com.vmware.vapi.std.errors.not_found": 404})
}


