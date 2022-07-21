package main

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"
)

func deleteSubject(c *cli.Context) error {
	record, kafkaTopic, versionStr := c.String(flagRecord.Name), c.String(flagTopicRequired.Name), c.String(flagVersion.Name)
	permanent := c.Bool(flagPermanent.Name)
	subject := subjectName(kafkaTopic, record)

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

	if version == 0 {
		err = schemaRegistryClient.DeleteSubject(subject, permanent)
	} else {
		err = schemaRegistryClient.DeleteSubjectByVersion(subject, version, permanent)
	}

	return err
}
