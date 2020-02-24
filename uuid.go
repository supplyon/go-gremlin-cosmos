package gremcos

import "github.com/gofrs/uuid"

// uuidGeneratorFunc is a function type for generating UUID's
type uuidGeneratorFunc func() (uuid.UUID, error)

// randomUUID is a function that randomly generates a new UUID
var randomUUID = func() (uuid.UUID, error) {
	return uuid.NewV4()
}
