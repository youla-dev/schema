package main

import (
	"fmt"
	"regexp"

	"github.com/urfave/cli/v2"
)

func subjects(c *cli.Context) error {
	kafkaTopic := c.String(flagTopicRequired.Name)
	r, err := regexp.Compile(fmt.Sprintf(`%s-\w+-value`, kafkaTopic))
	if err != nil {
		return fmt.Errorf("can not use topic name: %w", err)
	}

	schemaRegistryClient, err := getSRClient(c)
	if err != nil {
		return err
	}
	subjects, err := schemaRegistryClient.GetSubjects()
	if err != nil {
		return fmt.Errorf("can not get subjects: %w", err)
	}

	for _, subject := range subjects {
		if r.MatchString(subject) {
			fmt.Println(subject)
		}
	}

	return nil
}
