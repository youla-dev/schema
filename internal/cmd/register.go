package cmd

import (
	"context"
	"fmt"

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

func (r *Register) Run(ctx context.Context) (interface{}, error) {
	t := srclient.Protobuf
	var err error
	if r.topic == "" || r.record == "" {
		r.topic, r.record, err = protoschema.Parse(context.Background(), r.schemaBytes)
		if err != nil {
			return nil, fmt.Errorf("can not extract topic and record from proto: %w", err)
		}
	}

	if r.topic == "" {
		return nil, fmt.Errorf("topic %q is invalid", r.topic)
	}

	validatingSubject := subjectName(r.topic, r.record)

	// Topic search
	topics, err := r.clusterClient.Topics()
	if err != nil {
		return nil, fmt.Errorf("can not list topics: %w", err)
	}
	var topicExists bool
	for _, topic := range topics {
		if r.topic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		return fmt.Sprintf("topic %q not exist", r.topic), nil
	}

	// Does the topic exist
	subjects, err := r.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	schema, err := r.schemaRegistryClient.CreateSchema(validatingSubject, string(r.schemaBytes), t)
	if err != nil {
		return nil, fmt.Errorf("error creating the schema %w", err)
	}

	return fmt.Sprintln("Created schema ID", schema.ID(), ", version", schema.Version()), nil
}
