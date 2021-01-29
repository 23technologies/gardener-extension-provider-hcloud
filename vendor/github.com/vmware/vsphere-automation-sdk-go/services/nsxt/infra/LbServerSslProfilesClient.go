/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause */

// Code generated. DO NOT EDIT.

/*
 * Interface file for service: LbServerSslProfiles
 * Used by client-side stubs.
 */

package infra

import (
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

type LbServerSslProfilesClient interface {

    // Delete the LBServerSslProfile along with all the entities contained by this LBServerSslProfile.
    //
    // @param lbServerSslProfileIdParam LBServerSslProfile ID (required)
    // @param forceParam Force delete the resource even if it is being used somewhere (optional, default to false)
    // @throws InvalidRequest  Bad Request, Precondition Failed
    // @throws Unauthorized  Forbidden
    // @throws ServiceUnavailable  Service Unavailable
    // @throws InternalServerError  Internal Server Error
    // @throws NotFound  Not Found
	Delete(lbServerSslProfileIdParam string, forceParam *bool) error

    // Read a LBServerSslProfile.
    //
    // @param lbServerSslProfileIdParam LBServerSslProfile ID (required)
    // @return com.vmware.nsx_policy.model.LBServerSslProfile
    // @throws InvalidRequest  Bad Request, Precondition Failed
    // @throws Unauthorized  Forbidden
    // @throws ServiceUnavailable  Service Unavailable
    // @throws InternalServerError  Internal Server Error
    // @throws NotFound  Not Found
	Get(lbServerSslProfileIdParam string) (model.LBServerSslProfile, error)

    // Paginated list of all LBServerSslProfiles.
    //
    // @param cursorParam Opaque cursor to be used for getting next page of records (supplied by current result page) (optional)
    // @param includeMarkForDeleteObjectsParam Include objects that are marked for deletion in results (optional, default to false)
    // @param includedFieldsParam Comma separated list of fields that should be included in query result (optional)
    // @param pageSizeParam Maximum number of results to return in this page (server may return fewer) (optional, default to 1000)
    // @param sortAscendingParam (optional)
    // @param sortByParam Field by which records are sorted (optional)
    // @return com.vmware.nsx_policy.model.LBServerSslProfileListResult
    // @throws InvalidRequest  Bad Request, Precondition Failed
    // @throws Unauthorized  Forbidden
    // @throws ServiceUnavailable  Service Unavailable
    // @throws InternalServerError  Internal Server Error
    // @throws NotFound  Not Found
	List(cursorParam *string, includeMarkForDeleteObjectsParam *bool, includedFieldsParam *string, pageSizeParam *int64, sortAscendingParam *bool, sortByParam *string) (model.LBServerSslProfileListResult, error)

    // If a LBServerSslProfile with the lb-server-ssl-profile-id is not already present, create a new LBServerSslProfile. If it already exists, update the LBServerSslProfile. This is a full replace.
    //
    // @param lbServerSslProfileIdParam LBServerSslProfile ID (required)
    // @param lbServerSslProfileParam (required)
    // @throws InvalidRequest  Bad Request, Precondition Failed
    // @throws Unauthorized  Forbidden
    // @throws ServiceUnavailable  Service Unavailable
    // @throws InternalServerError  Internal Server Error
    // @throws NotFound  Not Found
	Patch(lbServerSslProfileIdParam string, lbServerSslProfileParam model.LBServerSslProfile) error

    // If a LBServerSslProfile with the lb-server-ssl-profile-id is not already present, create a new LBServerSslProfile. If it already exists, update the LBServerSslProfile. This is a full replace.
    //
    // @param lbServerSslProfileIdParam LBServerSslProfile ID (required)
    // @param lbServerSslProfileParam (required)
    // @return com.vmware.nsx_policy.model.LBServerSslProfile
    // @throws InvalidRequest  Bad Request, Precondition Failed
    // @throws Unauthorized  Forbidden
    // @throws ServiceUnavailable  Service Unavailable
    // @throws InternalServerError  Internal Server Error
    // @throws NotFound  Not Found
	Update(lbServerSslProfileIdParam string, lbServerSslProfileParam model.LBServerSslProfile) (model.LBServerSslProfile, error)
}
