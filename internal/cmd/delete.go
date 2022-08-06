package cmd

import (
	"context"

	"github.com/riferrei/srclient"
)

type Delete struct {
	schemaRegistryClient srclient.ISchemaRegistryClient
	topic, record        string
	version              int
	permanent            bool
}

func NewDelete(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic, record string,
	version int,
	permanent bool,
) (*Delete, error) {
	return &Delete{
		schemaRegistryClient: schemaRegistryClient,
		topic:                topic,
		record:               record,
		version:              version,
		permanent:            permanent,
	}, nil
}

func (d *Delete) Run(c context.Context) error {
	subject := subjectName(d.topic, d.record)
	var err error

	if d.version == 0 {
		err = d.schemaRegistryClient.DeleteSubject(subject, d.permanent)
	} else {
		err = d.schemaRegistryClient.DeleteSubjectByVersion(subject, d.version, d.permanent)
	}

	return err
}
