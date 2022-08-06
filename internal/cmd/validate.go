package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

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

func (v *Validate) Run(c context.Context) error {
	t := srclient.Protobuf
	var err error
	if v.topic == "" || v.record == "" {
		v.topic, v.record, err = protoschema.Parse(context.Background(), v.schemaBytes)
		if err != nil {
			return fmt.Errorf("can not extract topic and record from proto: %w", err)
		}
	}

	if v.topic == "" {
		log.Println("topic is not set", v.topic)
		return nil
	}

	validatingSubject := subjectName(v.topic, v.record)

	// Topic search
	topics, err := v.clusterClient.Topics()
	if err != nil {
		return fmt.Errorf("can not list topics: %w", err)
	}
	var topicExists bool
	for _, topic := range topics {
		if v.topic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		log.Printf("topic %q not exist", v.topic)
		return nil
	}

	subjects, err := v.schemaRegistryClient.GetSubjects()
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
	compatible, err := v.schemaRegistryClient.IsSchemaCompatible(validatingSubject, string(v.schemaBytes), "latest", t)
	if err != nil {
		return fmt.Errorf("error validating schema: %w", err)
	}
	if !compatible {
		return errors.New("schema is not compatible")
	}

	log.Println("schema is compatible")
	return nil
}
