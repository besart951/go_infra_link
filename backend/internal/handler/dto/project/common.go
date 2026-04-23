package project

import (
	commondto "github.com/besart951/go_infra_link/backend/internal/handler/dto/common"
	facilitydto "github.com/besart951/go_infra_link/backend/internal/handler/dto/facility"
)

type ErrorResponse = commondto.ErrorResponse
type PaginationQuery = commondto.PaginationQuery
type BacnetObjectResponse = facilitydto.BacnetObjectResponse
type ControlCabinetResponse = facilitydto.ControlCabinetResponse
type FieldDeviceOptionsResponse = facilitydto.FieldDeviceOptionsResponse
type ObjectDataListResponse = facilitydto.ObjectDataListResponse
type ObjectDataResponse = facilitydto.ObjectDataResponse
type SPSControllerResponse = facilitydto.SPSControllerResponse
type SPSControllerSystemTypeResponse = facilitydto.SPSControllerSystemTypeResponse
