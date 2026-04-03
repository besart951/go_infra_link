package project

type Handlers struct {
	Project            *ProjectHandler
	Phase              *PhaseHandler
	FieldDeviceOptions *FieldDeviceOptionsHandler
}

func NewHandlers(projectService ProjectService, phaseService PhaseService, fieldDeviceOptionsService FieldDeviceOptionsService) *Handlers {
	projectHandler := NewProjectHandler(projectService, nil)
	return &Handlers{
		Project:            projectHandler,
		Phase:              NewPhaseHandler(phaseService),
		FieldDeviceOptions: NewFieldDeviceOptionsHandler(projectService, fieldDeviceOptionsService),
	}
}
