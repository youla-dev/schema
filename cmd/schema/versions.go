package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

func versions(c *cli.Context) error {
	record, kafkaTopic := c.String(flagRecord.Name), c.String(flagTopicRequired.Name)
	validatingSubject := subjectName(kafkaTopic, record)

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
		return errors.New("schema not exist yet")
	}

	versions, err := schemaRegistryClient.GetSchemaVersions(validatingSubject)
	if err != nil {
		return fmt.Errorf("error getting versions: %w", err)
	}

	for _, version := range versions {
		fmt.Println(version)
	}
	return nil
}
