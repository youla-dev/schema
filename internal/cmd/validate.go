package cmd

import (
	"context"
	"errors"
	"fmt"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/riferrei/srclient"
	"github.com/youla-dev/schema/lib/protoschema"
)

type Validate struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	clusterClient        *saramaCluster.Client
	topic, record        string
	schemaBytes          []byte
}

func NewValidate(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	clusterClient *saramaCluster.Client,
	topic, record string,
	schemaBytes []byte,
) (*Validate, error) {
	return &Validate{
		schemaRegistryClient: schemaRegistryClient,
		clusterClient:        clusterClient,
		topic:                topic,
		record:               record,
		schemaBytes:          schemaBytes,
	}, nil
}

func (v *Validate) Run(c context.Context) (interface{}, error) {
	t := srclient.Protobuf
	var err error
	if v.topic == "" || v.record == "" {
		v.topic, v.record, err = protoschema.Parse(context.Background(), v.schemaBytes)
		if err != nil {
			return nil, fmt.Errorf("can not extract topic and record from proto: %w", err)
		}
	}

	if v.topic == "" {
		return nil, fmt.Errorf("topic %q is invalid", v.topic)
	}

	validatingSubject := subjectName(v.topic, v.record)

	// Topic search
	topics, err := v.clusterClient.Topics()
	if err != nil {
		return nil, fmt.Errorf("can not list topics: %w", err)
	}
	var topicExists bool
	for _, topic := range topics {
		if v.topic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		return fmt.Sprintf("topic %q not exist", v.topic), nil
	}

	subjects, err := v.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}
	var subjectExist bool
	for _, subject := range subjects {
		subjectExist = subjectExist || subject == validatingSubject
	}

	if !subjectExist {
		return fmt.Sprintf("schema %q not exist yet", validatingSubject), nil
	}

	// Is the scheme compatible
	compatible, err := v.schemaRegistryClient.IsSchemaCompatible(validatingSubject, string(v.schemaBytes), "latest", t)
	if err != nil {
		return nil, fmt.Errorf("error validating schema: %w", err)
	}
	if !compatible {
		return nil, errors.New("schema is not compatible")
	}

	return "schema is compatible", nil
}
