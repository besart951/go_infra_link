package facility

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type apparatNumberScope struct {
	systemTypeID uuid.UUID
	systemPartID uuid.UUID
	apparatID    uuid.UUID
}

type fieldDeviceCreateWorkItem struct {
	index int
	item  domainFacility.FieldDeviceCreateItem
}

type fieldDeviceCreateExecution struct {
	ctx    context.Context
	writer fieldDeviceWriter
	work   []fieldDeviceCreateWorkItem
	result *domainFacility.FieldDeviceMultiCreateResult
}

type fieldDeviceCreateProblem struct {
	field   string
	message string
	cause   error
}

type fieldDeviceCreatePlanner struct {
	ctx         context.Context
	service     *FieldDeviceService
	systemTypes map[uuid.UUID]*domainFacility.SPSControllerSystemType
	knownTypes  map[uuid.UUID]bool
	apparats    map[uuid.UUID]bool
	systemParts map[uuid.UUID]bool
	numbers     map[apparatNumberScope]map[int]struct{}
}

func newFieldDeviceCreatePlanner(ctx context.Context, service *FieldDeviceService) *fieldDeviceCreatePlanner {
	return &fieldDeviceCreatePlanner{
		ctx: ctx, service: service,
		systemTypes: make(map[uuid.UUID]*domainFacility.SPSControllerSystemType),
		knownTypes:  make(map[uuid.UUID]bool), apparats: make(map[uuid.UUID]bool),
		systemParts: make(map[uuid.UUID]bool), numbers: make(map[apparatNumberScope]map[int]struct{}),
	}
}

func (p *fieldDeviceCreatePlanner) prepare(items []domainFacility.FieldDeviceCreateItem, result *domainFacility.FieldDeviceMultiCreateResult) []fieldDeviceCreateWorkItem {
	work := make([]fieldDeviceCreateWorkItem, 0, len(items))
	for index, item := range items {
		resultItem := &result.Results[index]
		resultItem.Index = index
		problem := p.validate(item)
		if problem != nil {
			applyFieldDeviceCreateProblem(resultItem, problem)
			result.FailureCount++
			continue
		}
		work = append(work, fieldDeviceCreateWorkItem{index: index, item: item})
	}
	return work
}

func (p *fieldDeviceCreatePlanner) validate(item domainFacility.FieldDeviceCreateItem) *fieldDeviceCreateProblem {
	if item.FieldDevice == nil {
		return createProblem("fielddevice", "field device is required", nil)
	}
	if item.ObjectDataID != nil && len(item.BacnetObjects) > 0 {
		return createProblem("fielddevice", "object_data_id and bacnet_objects are mutually exclusive", nil)
	}
	if err := p.service.validateRequiredFields(item.FieldDevice); err != nil {
		return createProblem("fielddevice", "", err)
	}
	if err := p.ensureParents(item.FieldDevice); err != nil {
		return parentCreateProblem(err)
	}
	if problem := validateApparatNumber(item.FieldDevice.ApparatNr); problem != nil {
		return problem
	}
	return p.reserveApparatNumber(item.FieldDevice)
}

func (p *fieldDeviceCreatePlanner) ensureParents(device *domainFacility.FieldDevice) error {
	systemType, err := p.loadControllerSystemType(device.SPSControllerSystemTypeID)
	if err != nil {
		return err
	}
	if err := p.ensureSystemType(systemType.SystemTypeID); err != nil {
		return err
	}
	if err := p.ensureApparat(device.ApparatID); err != nil {
		return err
	}
	return p.ensureSystemPart(device.SystemPartID)
}

func (p *fieldDeviceCreatePlanner) loadControllerSystemType(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	if cached, exists := p.systemTypes[id]; exists {
		if cached == nil {
			return nil, domain.ErrNotFound
		}
		return cached, nil
	}
	item, err := domain.GetByID(p.ctx, p.service.spsControllerSystemTypeRepo, id)
	if err != nil {
		p.systemTypes[id] = nil
		return nil, err
	}
	p.systemTypes[id] = item
	return item, nil
}

func (p *fieldDeviceCreatePlanner) ensureSystemType(id uuid.UUID) error {
	if p.knownTypes[id] {
		return nil
	}
	if _, err := domain.GetByID(p.ctx, p.service.systemTypeRepo, id); err != nil {
		return err
	}
	p.knownTypes[id] = true
	return nil
}

func (p *fieldDeviceCreatePlanner) ensureApparat(id uuid.UUID) error {
	if p.apparats[id] {
		return nil
	}
	if _, err := domain.GetByID(p.ctx, p.service.apparatRepo, id); err != nil {
		return err
	}
	p.apparats[id] = true
	return nil
}

func (p *fieldDeviceCreatePlanner) ensureSystemPart(id uuid.UUID) error {
	if p.systemParts[id] {
		return nil
	}
	if _, err := domain.GetByID(p.ctx, p.service.systemPartRepo, id); err != nil {
		return err
	}
	p.systemParts[id] = true
	return nil
}

func (p *fieldDeviceCreatePlanner) reserveApparatNumber(device *domainFacility.FieldDevice) *fieldDeviceCreateProblem {
	scope := apparatNumberScope{
		systemTypeID: device.SPSControllerSystemTypeID,
		systemPartID: device.SystemPartID,
		apparatID:    device.ApparatID,
	}
	used, err := p.usedApparatNumbers(scope)
	if err != nil {
		return createProblem("fielddevice.apparat_nr", "", err)
	}
	if _, exists := used[device.ApparatNr]; exists {
		return createProblem("fielddevice.apparat_nr", apparatNrAlreadyUsedMessage, nil)
	}
	used[device.ApparatNr] = struct{}{}
	return nil
}

func (p *fieldDeviceCreatePlanner) usedApparatNumbers(scope apparatNumberScope) (map[int]struct{}, error) {
	if used, exists := p.numbers[scope]; exists {
		return used, nil
	}
	numbers, err := p.service.repo.GetUsedApparatNumbers(
		p.ctx, scope.systemTypeID, scope.systemPartID, scope.apparatID,
	)
	if err != nil {
		return nil, err
	}
	used := make(map[int]struct{}, len(numbers))
	for _, number := range numbers {
		used[number] = struct{}{}
	}
	p.numbers[scope] = used
	return used, nil
}

func executeFieldDeviceCreateWork(execution fieldDeviceCreateExecution) {
	for _, candidate := range execution.work {
		resultItem := &execution.result.Results[candidate.index]
		selection := fieldDeviceBacnetSelection{
			objectDataID: candidate.item.ObjectDataID, objects: candidate.item.BacnetObjects,
			objectsSet: len(candidate.item.BacnetObjects) > 0,
		}
		if err := execution.writer.create(execution.ctx, candidate.item.FieldDevice, selection); err != nil {
			setFieldDeviceCreateError(resultItem, err, "fielddevice")
			execution.result.FailureCount++
			continue
		}
		resultItem.Success = true
		resultItem.FieldDevice = candidate.item.FieldDevice
		execution.result.SuccessCount++
	}
}

func validateApparatNumber(number int) *fieldDeviceCreateProblem {
	if number == 0 {
		return createProblem("fielddevice.apparat_nr", "apparat_nr is required", nil)
	}
	if number < 1 || number > 99 {
		return createProblem("fielddevice.apparat_nr", "apparat_nr must be between 1 and 99", nil)
	}
	return nil
}

func parentCreateProblem(err error) *fieldDeviceCreateProblem {
	if errors.Is(err, domain.ErrNotFound) {
		return createProblem("fielddevice", "one or more parent entities (SPS controller, apparat, system part) not found", nil)
	}
	return createProblem("fielddevice", "", err)
}

func createProblem(field, message string, cause error) *fieldDeviceCreateProblem {
	return &fieldDeviceCreateProblem{field: field, message: message, cause: cause}
}

func applyFieldDeviceCreateProblem(result *domainFacility.FieldDeviceCreateResult, problem *fieldDeviceCreateProblem) {
	result.Success = false
	result.ErrorField = problem.field
	if problem.cause != nil {
		setFieldDeviceCreateError(result, problem.cause, problem.field)
		return
	}
	result.Error = problem.message
}
