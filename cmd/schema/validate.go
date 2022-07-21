package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
	"github.com/youla-dev/schema/lib/protoschema"
)

func validate(c *cli.Context) error {
	record, kafkaTopic := c.String(flagRecord.Name), c.String(flagTopic.Name)
	protoFile := c.String(flagProtoRequired.Name)

	t := srclient.Protobuf
	schemaBytes, err := os.ReadFile(protoFile)
	if err != nil {
		return fmt.Errorf("error reading schema: %w", err)
	}
	if kafkaTopic == "" || record == "" {
		kafkaTopic, record, err = protoschema.Parse(context.Background(), schemaBytes)
		if err != nil {
			return fmt.Errorf("can not extract topic and record from proto: %w", err)
		}
	}

	if kafkaTopic == "" {
		log.Println("topic is not set", kafkaTopic)
		return nil
	}

	validatingSubject := subjectName(kafkaTopic, record)

	// Topic search
	clusterClient, err := getClusterClient(c)
	if err != nil {
		return err
	}
	topics, err := clusterClient.Topics()
	if err != nil {
		return fmt.Errorf("can not list topics: %w", err)
	}
	var topicExists bool
	for _, topic := range topics {
		if kafkaTopic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		log.Printf("topic %q not exist", kafkaTopic)
		return nil
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
		log.Printf("schema %q not exist yet", validatingSubject)
		return nil
	}

	// Is the scheme compatible
	compatible, err := schemaRegistryClient.IsSchemaCompatible(validatingSubject, string(schemaBytes), "latest", t)
	if err != nil {
		return fmt.Errorf("error validating schema: %w", err)
	}
	if !compatible {
		return errors.New("schema is not compatible")
	}

	log.Println("schema is compatible")
	return nil
}
