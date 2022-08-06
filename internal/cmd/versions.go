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

func (v *Versions) Run(c context.Context) error {
	validatingSubject := subjectName(v.topic, v.record)

	subjects, err := v.schemaRegistryClient.GetSubjects()
	if err != nil {
		return fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		return errors.New("schema not exist yet")
	}

	versions, err := v.schemaRegistryClient.GetSchemaVersions(validatingSubject)
	if err != nil {
		return fmt.Errorf("error getting versions: %w", err)
	}

	for _, version := range versions {
		fmt.Println(version)
	}
	return nil
}
