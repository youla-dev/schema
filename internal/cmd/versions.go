package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/riferrei/srclient"
)

type Versions struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	topic, record        string
}

func NewVersions(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic, record string,
) (*Versions, error) {
	return &Versions{
		schemaRegistryClient: schemaRegistryClient,
		topic:                topic,
		record:               record,
	}, nil
}

func (v *Versions) Run(c context.Context) (interface{}, error) {
	validatingSubject := subjectName(v.topic, v.record)

	subjects, err := v.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		return nil, errors.New("schema not exist yet")
	}

	versions, err := v.schemaRegistryClient.GetSchemaVersions(validatingSubject)
	if err != nil {
		return nil, fmt.Errorf("error getting versions: %w", err)
	}
	return versions, nil
}
