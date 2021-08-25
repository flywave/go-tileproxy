package utils

import (
	"encoding/binary"
	"time"

	"github.com/gofrs/uuid"
)

var (
	constantUUID = uuid.Must(uuid.FromString("e8dba9f7-21e2-4c82-96cb-6586922c6422"))
	instanceUUID = RandomUUID("instance")
)

func RandomUUID(ns string) uuid.UUID {
	randUUID, err := uuid.NewV4()
	switch {
	case err != nil:
		return uuid.NewV5(uuidFromTime(), ns)
	case ns != "":
		return uuid.NewV5(randUUID, ns)
	default:
		return randUUID
	}
}

func DerivedUUID(input string) uuid.UUID {
	return uuid.NewV5(constantUUID, input)
}

func DerivedInstanceUUID(input string) uuid.UUID {
	return uuid.NewV5(instanceUUID, input)
}

func uuidFromTime() uuid.UUID {
	var timeUUID uuid.UUID
	binary.LittleEndian.PutUint64(timeUUID[:], uint64(time.Now().UnixNano()))
	return timeUUID
}
