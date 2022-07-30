package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pterm/pterm"
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

func (i *Inspect) Run(c context.Context) error {
	validatingSubject := subjectName(i.topic, i.record)

	subjects, err := i.schemaRegistryClient.GetSubjects()
	if err != nil {
		return fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		log.Println("schema not exist yet")
		return nil
	}

	var schema *srclient.Schema
	if i.version == 0 {
		schema, err = i.schemaRegistryClient.GetLatestSchema(validatingSubject)
	} else {
		schema, err = i.schemaRegistryClient.GetSchemaByVersion(validatingSubject, i.version)
	}
	if err != nil {
		return fmt.Errorf("error schema: %w", err)
	}

	references, err := json.Marshal(schema.References())
	if err != nil {
		return fmt.Errorf("can not marshal references: %w", err)
	}

	tbl := pterm.DefaultTable.WithData(pterm.TableData{
		{"Subject", validatingSubject},
		{"ID", strconv.Itoa(schema.ID())},
		{"Version", strconv.Itoa(schema.Version())},
		{"References", string(references)},
		{"Schema:"},
	})
	if err := tbl.Render(); err != nil {
		return fmt.Errorf("can not render: %w", err)
	}

	fmt.Println()
	fmt.Println(schema.Schema())

	return nil
}
