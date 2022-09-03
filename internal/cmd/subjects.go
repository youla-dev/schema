package cmd

import (
	"context"
	"fmt"
	"regexp"

	"github.com/riferrei/srclient"
)

type Subjects struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	topic                string
}

func NewSubjects(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic string,
) (*Subjects, error) {
	return &Subjects{
		schemaRegistryClient: schemaRegistryClient,
		topic:                topic,
	}, nil
}

func (s *Subjects) Run(c context.Context) (interface{}, error) {
	r, err := regexp.Compile(fmt.Sprintf(`%s-\w+-value`, s.topic))
	if err != nil {
		return nil, fmt.Errorf("can not use topic name: %w", err)
	}

	subjects, err := s.schemaRegistryClient.GetSubjects()
	if err != nil {
		return nil, fmt.Errorf("can not get subjects: %w", err)
	}

	filtered := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		if r.MatchString(subject) {
			filtered = append(filtered, subject)
		}
	}

	return filtered, nil
}
