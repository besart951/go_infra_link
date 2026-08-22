package facilitybenchmark

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ProjectCount     = 100
	CabinetCount     = 1_000
	ControllerCount  = 10_000
	SystemScopeCount = 50_506
	FieldDeviceCount = 5_000_000

	searchTokenPointOnePercent = "selectivity_p001"
	searchTokenOnePercent      = "selectivity_p01"
	searchTokenTenPercent      = "selectivity_p10"
)

var benchmarkEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func deterministicID(kind byte, index int64) uuid.UUID {
	var id uuid.UUID
	id[0] = kind
	binary.BigEndian.PutUint64(id[8:], uint64(index+1))
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}

type generatedSource struct {
	count int64
	index int64
	row   func(int64) ([]any, error)
	err   error
}

func (s *generatedSource) Next() bool { return s.err == nil && s.index < s.count }

func (s *generatedSource) Values() ([]any, error) {
	values, err := s.row(s.index)
	s.index++
	return values, err
}

func (s *generatedSource) Err() error { return s.err }

func textValue(index int64) string {
	markers := searchTokens(index)
	if len(markers) == 0 {
		return fmt.Sprintf("field device %d", index)
	}
	return fmt.Sprintf("%s device %d", strings.Join(markers, " "), index)
}

func searchTokens(index int64) []string {
	markers := make([]string, 0, 3)
	if index%1_000 == 0 {
		markers = append(markers, searchTokenPointOnePercent)
	}
	if index%100 == 0 {
		markers = append(markers, searchTokenOnePercent)
	}
	if index%10 == 0 {
		markers = append(markers, searchTokenTenPercent)
	}
	return markers
}
