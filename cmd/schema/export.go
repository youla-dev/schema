package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
)

func export(c *cli.Context) error {
	record, kafkaTopic, versionStr := c.String(flagRecord.Name), c.String(flagTopicRequired.Name), c.String(flagVersion.Name)
	output := c.String("output")
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
		fmt.Println("schema not exist yet")
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

	if err := os.WriteFile(output, []byte(schema.Schema()), 0666); err != nil {
		return fmt.Errorf("can not write scheme to the file: %w", err)
	}

	return nil
}
