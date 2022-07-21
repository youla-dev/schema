package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
)

func inspect(c *cli.Context) error {
	record, kafkaTopic, versionStr := c.String(flagRecord.Name), c.String(flagTopicRequired.Name), c.String(flagVersion.Name)
	validatingSubject := subjectName(kafkaTopic, record)

	var version int
	if versionStr != "latest" {
		var err error
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("version is invalid: %w", err)
		}
	}

	schemaRegistryClient, err := getSRClient(c)
	if err != nil {
		return err
	}
	subjects, err := schemaRegistryClient.GetSubjects()
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
	if version == 0 {
		schema, err = schemaRegistryClient.GetLatestSchema(validatingSubject)
	} else {
		schema, err = schemaRegistryClient.GetSchemaByVersion(validatingSubject, version)
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
