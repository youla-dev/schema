package cmd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/riferrei/srclient"
)

type Export struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	topic, record        string
	version              int
}

func NewExport(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic, record string,
	version int,
) (*Export, error) {
	return &Export{
		schemaRegistryClient: schemaRegistryClient,
		topic:                topic,
		record:               record,
		version:              version,
	}, nil
}

func (e *Export) Run(c context.Context) (interface{}, error) {
	validatingSubject := subjectName(e.topic, e.record)

	subjects, err := e.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		fmt.Println("schema not exist yet")
		return nil, nil
	}

	var schema *srclient.Schema
	if e.version == 0 {
		schema, err = e.schemaRegistryClient.GetLatestSchema(validatingSubject)
	} else {
		schema, err = e.schemaRegistryClient.GetSchemaByVersion(validatingSubject, e.version)
	}
	if err != nil {
		return nil, fmt.Errorf("error schema: %w", err)
	}

	return bytes.NewBufferString(schema.Schema()), nil
}
