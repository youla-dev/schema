package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
	"github.com/youla-dev/schema/internal/cmd"
)

const (
	cmdRegister = "register"
	cmdDelete   = "delete"
	cmdValidate = "validate"
	cmdVersions = "versions"
	cmdInspect  = "inspect"
	cmdSubjects = "subjects"
	cmdExport   = "export"
)

var (
	FlagClusterRequired = &cli.StringSliceFlag{
		Name:     "cluster",
		Required: true,
		Value:    cli.NewStringSlice("localhost:9092"),
		Usage:    "List of Kafka brokers joined with comma.",
		EnvVars:  []string{"CLUSTER"},
	}
	FlagSRRequired = &cli.StringFlag{
		Name:     "sr",
		Required: true,
		Value:    "http://localhost:8081",
		Usage:    "URL for the Confluent Schema Registry.",
		EnvVars:  []string{"SCHEMA_REGISTRY"},
	}
	FlagSR = &cli.StringFlag{
		Name:    "sr",
		Value:   "http://localhost:8081",
		Usage:   "URL for the Confluent Schema Registry.",
		EnvVars: []string{"SCHEMA_REGISTRY"},
	}
	FlagTopicRequired = &cli.StringFlag{
		Name:     "topic",
		Required: true,
		Value:    "test-topic",
		Usage:    "Name of the topic.",
		EnvVars:  []string{"TOPIC"},
	}
	FlagTopic = &cli.StringFlag{
		Name:    "topic",
		Usage:   "Name of the topic.",
		EnvVars: []string{"TOPIC"},
	}
	FlagRecord = &cli.StringFlag{
		Name:    "record",
		Value:   "default",
		Usage:   "Name of the record within the topic.",
		EnvVars: []string{"RECORD"},
	}
	FlagPermanent = &cli.BoolFlag{
		Name:    "permanent",
		Value:   false,
		Usage:   "Permanent flag is used for the removing of version in 'hard' mode.",
		EnvVars: []string{"PERMANENT"},
	}
	FlagProtoRequired = &cli.StringFlag{
		Name:     "proto",
		Required: true,
		Usage:    "File with Protobuf schema definition.",
		EnvVars:  []string{"PROTO"},
	}
	FlagVersion = &cli.StringFlag{
		Name:    "version",
		Value:   "latest",
		Usage:   "Version of the schema. E.g. `latest`, `1`, etc.",
		EnvVars: []string{"VERSION"},
	}
)

type App struct {
	cliApp *cli.App
}

func NewApp(version string) *App {
	app := &App{
		cliApp: &cli.App{
			Name:    "Schema Registry utility",
			Version: version,
			Usage:   "Utility for your CI/CD process to validate, register or delete Kafka protobuf schemes in the registry.",
		},
	}
	app.cliApp.Commands = []*cli.Command{
		{
			Name:      cmdRegister,
			Usage:     "Creates a subject, if one does not exists, sets a scheme for a subject or updates it.",
			ArgsUsage: "Set both the kafka cluster and the registry addresses and the protobuf file with the schema and topic&record options or set topic and record values with the separate flags.",
			Action:    app.register,
			Flags: []cli.Flag{
				FlagClusterRequired,
				FlagSRRequired,
				FlagTopic,
				FlagRecord,
				FlagProtoRequired,
			},
		},
		{
			Name:      cmdDelete,
			Usage:     "Deletes the specified version of the schema or deletes the latest version of the scheme by default.",
			ArgsUsage: "Set the topic, record and version of the schema to delete.",
			Action:    app.delete,
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
				FlagVersion,
				FlagPermanent,
			},
		},
		{
			Name:   cmdValidate,
			Usage:  "Validates the topic to exist and the schema changes compatibility with existing version. The schema is also valid if the the topic or subject does not exists.",
			Action: app.validate,
			Flags: []cli.Flag{
				FlagClusterRequired,
				FlagSRRequired,
				FlagTopic,
				FlagRecord,
				FlagProtoRequired,
			},
		},
		{
			Name:   cmdVersions,
			Usage:  "Lists available versions for the subject.",
			Action: app.versions,
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
			},
		},
		{
			Name:   cmdInspect,
			Usage:  "Outputs all information about the subject.",
			Action: app.inspect,
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
				FlagVersion,
			},
		},
		{
			Name:   cmdSubjects,
			Usage:  "Lists available records for the topic.",
			Action: app.subjects,
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
			},
		},
		{
			Name:   cmdExport,
			Usage:  "Exports the schema value to the local file.",
			Action: app.export,
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
				FlagVersion,
				&cli.StringFlag{
					Name:     "output",
					Required: true,
					Usage:    "Output file for the scheme from registry.",
					EnvVars:  []string{"OUTPUT"},
				},
			},
		},
	}

	return app
}

func (a *App) Run(ctx context.Context) {
	err := a.cliApp.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func (a *App) inspect(c *cli.Context) error {
	record, topic, versionStr := c.String(FlagRecord.Name), c.String(FlagTopicRequired.Name), c.String(FlagVersion.Name)
	var version int
	if versionStr != "latest" {
		var err error
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("version is invalid: %w", err)
		}
	}
	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}
	i, err := cmd.NewInspect(schemaRegistryClient, topic, record, version)
	if err != nil {
		return err
	}
	return i.Run(c.Context)
}

func (a *App) register(c *cli.Context) error {
	record, topic := c.String(FlagRecord.Name), c.String(FlagTopic.Name)
	protoFile := c.String(FlagProtoRequired.Name)

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	clusterClient, err := a.getClusterClient(c)
	if err != nil {
		return err
	}

	schemaBytes, err := os.ReadFile(protoFile)
	if err != nil {
		return fmt.Errorf("error reading schema: %w", err)
	}

	r, err := cmd.NewRegister(schemaRegistryClient, clusterClient, topic, record, schemaBytes)
	if err != nil {
		return err
	}
	return r.Run(c.Context)
}

func (a *App) validate(c *cli.Context) error {
	record, topic := c.String(FlagRecord.Name), c.String(FlagTopic.Name)
	protoFile := c.String(FlagProtoRequired.Name)

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	clusterClient, err := a.getClusterClient(c)
	if err != nil {
		return err
	}

	schemaBytes, err := os.ReadFile(protoFile)
	if err != nil {
		return fmt.Errorf("error reading schema: %w", err)
	}

	v, err := cmd.NewValidate(schemaRegistryClient, clusterClient, topic, record, schemaBytes)
	if err != nil {
		return err
	}
	return v.Run(c.Context)
}

func (a *App) versions(c *cli.Context) error {
	record, topic := c.String(FlagRecord.Name), c.String(FlagTopic.Name)

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	v, err := cmd.NewVersions(schemaRegistryClient, topic, record)
	if err != nil {
		return err
	}
	return v.Run(c.Context)
}

func (a *App) delete(c *cli.Context) error {
	record, topic, versionStr := c.String(FlagRecord.Name), c.String(FlagTopicRequired.Name), c.String(FlagVersion.Name)
	permanent := c.Bool(FlagPermanent.Name)

	var version int
	if versionStr != "latest" {
		var err error
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("version is invalid: %w", err)
		}
	}

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	d, err := cmd.NewDelete(schemaRegistryClient, topic, record, version, permanent)
	if err != nil {
		return err
	}
	return d.Run(c.Context)
}

func (a *App) subjects(c *cli.Context) error {
	topic := c.String(FlagTopicRequired.Name)

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	s, err := cmd.NewSubjects(schemaRegistryClient, topic)
	if err != nil {
		return err
	}
	return s.Run(c.Context)
}

func (a *App) export(c *cli.Context) error {
	record, topic, versionStr := c.String(FlagRecord.Name), c.String(FlagTopicRequired.Name), c.String(FlagVersion.Name)

	var version int
	if versionStr != "latest" {
		var err error
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("version is invalid: %w", err)
		}
	}

	schemaRegistryClient, err := a.getSRClient(c)
	if err != nil {
		return err
	}

	e, err := cmd.NewExport(schemaRegistryClient, topic, record, version)
	if err != nil {
		return err
	}
	output, err := e.Run(c.Context)
	if err != nil {
		return err
	}

	outputFile := c.String("output")
	if err := os.WriteFile(outputFile, output, 0666); err != nil {
		return fmt.Errorf("can not write scheme to the file: %w", err)
	}
	return nil
}

func (a *App) getSRClient(c *cli.Context) (srclient.ISchemaRegistryClient, error) {
	return srclient.CreateSchemaRegistryClient(c.String(FlagSRRequired.Name)), nil
}

func (a *App) getClusterClient(c *cli.Context) (*saramaCluster.Client, error) {
	cluster := c.StringSlice(FlagClusterRequired.Name)
	kfkCfg := saramaCluster.NewConfig()
	clusterClient, err := saramaCluster.NewClient(cluster, kfkCfg)
	if err != nil {
		return nil, fmt.Errorf("can not create cluster client %v: %w", cluster, err)
	}
	return clusterClient, nil
}

func subjectName(kafkaTopic, record string) string {
	return kafkaTopic + "-" + record + "-value"
}
