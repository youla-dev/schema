package cmd

import (
	"context"
	"fmt"
	"log"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/riferrei/srclient"
	"github.com/youla-dev/schema/lib/protoschema"
)

type Register struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	clusterClient        *saramaCluster.Client
	topic, record        string
	schemaBytes          []byte
}

func NewRegister(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	clusterClient *saramaCluster.Client,
	topic, record string,
	schemaBytes []byte,
) (*Register, error) {
	return &Register{
		schemaRegistryClient: schemaRegistryClient,
		clusterClient:        clusterClient,
		topic:                topic,
		record:               record,
		schemaBytes:          schemaBytes,
	}, nil
}

func (r *Register) Run(ctx context.Context) error {
	t := srclient.Protobuf
	var err error
	if r.topic == "" || r.record == "" {
		r.topic, r.record, err = protoschema.Parse(context.Background(), r.schemaBytes)
		if err != nil {
			return fmt.Errorf("can not extract topic and record from proto: %w", err)
		}
	}

	if r.topic == "" {
		log.Println("topic is not set", r.topic)
		return nil
	}

	validatingSubject := subjectName(r.topic, r.record)

	// Topic search
	topics, err := r.clusterClient.Topics()
	if err != nil {
		return fmt.Errorf("can not list topics: %w", err)
	}
	var topicExists bool
	for _, topic := range topics {
		if r.topic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		log.Printf("topic %q not exist", r.topic)
		return nil
	}

	// Does the topic exist

	subjects, err := r.schemaRegistryClient.GetSubjects()
	if err != nil {
		return fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	schema, err := r.schemaRegistryClient.CreateSchema(validatingSubject, string(r.schemaBytes), t)
	if err != nil {
		return fmt.Errorf("error creating the schema %w", err)
	}

	log.Println("Created schema ID", schema.ID(), ", version", schema.Version())

	return nil
}
