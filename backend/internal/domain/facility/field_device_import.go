package facility

// FieldDeviceImportAggregate is the lossless, owned graph accepted by the
// import command. Shared reference data is referenced by ID and never copied.
type FieldDeviceImportAggregate struct {
	FieldDevice   FieldDevice
	Specification *Specification
	BacnetObjects []BacnetObject
}
