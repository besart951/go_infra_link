package facilitysql

import "github.com/google/uuid"

func argStringPtr(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func argIntPtr(i *int) any {
	if i == nil {
		return nil
	}
	return int64(*i)
}

func argFloatPtr(f *float64) any {
	if f == nil {
		return nil
	}
	return *f
}

func argUUIDPtr(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}

func ToNull[T any](ptr *T) any {
	if ptr == nil {
		return nil
	}
	return *ptr
}
