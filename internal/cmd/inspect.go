package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riferrei/srclient"
)

type Inspect struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	topic, record        string
	version              int
}

func NewInspect(schemaRegistryClient srclient.ISchemaRegistryClient, topic, record string, version int) (*Inspect, error) {
	return &Inspect{
		schemaRegistryClient: schemaRegistryClient,
		topic:                topic,
		record:               record,
		version:              version,
	}, nil
}

type inspectOutput struct {
	Subject    string `json:"subject"`
	ID         int    `json:"id"`
	Version    int    `json:"version"`
	References string `json:"references"`
	Schema     string `json:"schema"`
}

func (i *Inspect) Run(c context.Context) (interface{}, error) {
	validatingSubject := subjectName(i.topic, i.record)

	subjects, err := i.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		return "schema not exist yet", nil
	}

	var schema *srclient.Schema
	if i.version == 0 {
		schema, err = i.schemaRegistryClient.GetLatestSchema(validatingSubject)
	} else {
		schema, err = i.schemaRegistryClient.GetSchemaByVersion(validatingSubject, i.version)
	}
	if err != nil {
		return nil, fmt.Errorf("error schema: %w", err)
	}

	references, err := json.Marshal(schema.References())
	if err != nil {
		return nil, fmt.Errorf("can not marshal references: %w", err)
	}

	output := inspectOutput{
		Subject:    validatingSubject,
		ID:         schema.ID(),
		Version:    schema.Version(),
		References: string(references),
		Schema:     schema.Schema(),
	}

	return output, nil
}
