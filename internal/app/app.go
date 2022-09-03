package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	saramaCluster "github.com/bsm/sarama-cluster"
	"github.com/riferrei/srclient"
	"github.com/urfave/cli/v2"
	"github.com/youla-dev/schema/internal/cmd"
	"go.uber.org/dig"
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
	FlagOutputRequired = &cli.StringFlag{
		Name:     "output",
		Required: true,
		Usage:    "Output file for the scheme from registry.",
		EnvVars:  []string{"OUTPUT"},
	}
)

type (
	ClusterFlag   []string
	SRFlag        string
	TopicFlag     string
	RecordFlag    string
	PermanentFlag bool
	ProtoFlag     string
	VersionFlag   int
	OutputFlag    string
)

func GetClusterFlag(c *cli.Context) ClusterFlag {
	return ClusterFlag(c.StringSlice(FlagClusterRequired.Name))
}

func GetSRFlag(c *cli.Context) SRFlag {
	return SRFlag(c.String(FlagSRRequired.Name))
}

func GetTopicFlag(c *cli.Context) TopicFlag {
	return TopicFlag(c.String(FlagTopicRequired.Name))
}

func GetRecordFlag(c *cli.Context) RecordFlag {
	return RecordFlag(c.String(FlagRecord.Name))
}

func GetPermanentFlag(c *cli.Context) PermanentFlag {
	return PermanentFlag(c.Bool(FlagPermanent.Name))
}

func GetProtoFlag(c *cli.Context) ProtoFlag {
	return ProtoFlag(c.String(FlagProtoRequired.Name))
}

func GetVersionFlag(c *cli.Context) (VersionFlag, error) {
	versionStr := c.String(FlagVersion.Name)
	var version int
	if versionStr != "latest" {
		var err error
		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return 0, fmt.Errorf("version is invalid: %w", err)
		}
	}
	return VersionFlag(version), nil
}

func GetOutputFlag(c *cli.Context) OutputFlag {
	return OutputFlag(c.String(FlagOutputRequired.Name))
}

func GetClusterClient(connection ClusterFlag) (*saramaCluster.Client, error) {
	kfkCfg := saramaCluster.NewConfig()
	clusterClient, err := saramaCluster.NewClient(connection, kfkCfg)
	if err != nil {
		return nil, fmt.Errorf("can not create cluster client %v: %w", connection, err)
	}
	return clusterClient, nil
}

func GetSRClient(connection SRFlag) srclient.ISchemaRegistryClient {
	return srclient.CreateSchemaRegistryClient(string(connection))
}

func GetInspect(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
	version VersionFlag,
) (*cmd.Inspect, error) {
	return cmd.NewInspect(schemaRegistryClient, string(topic), string(record), int(version))
}

func GetRegister(
	clusterClient *saramaCluster.Client,
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
	protoFile ProtoFlag,
) (*cmd.Register, error) {
	schemaBytes, err := os.ReadFile(string(protoFile))
	if err != nil {
		return nil, fmt.Errorf("error reading schema: %w", err)
	}
	return cmd.NewRegister(schemaRegistryClient, clusterClient, string(topic), string(record), schemaBytes)
}

func GetValidate(
	clusterClient *saramaCluster.Client,
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
	protoFile ProtoFlag,
) (*cmd.Validate, error) {
	schemaBytes, err := os.ReadFile(string(protoFile))
	if err != nil {
		return nil, fmt.Errorf("error reading schema: %w", err)
	}
	return cmd.NewValidate(schemaRegistryClient, clusterClient, string(topic), string(record), schemaBytes)
}

func GetDelete(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
	version VersionFlag,
	permanent PermanentFlag,
) (*cmd.Delete, error) {
	return cmd.NewDelete(schemaRegistryClient, string(topic), string(record), int(version), bool(permanent))
}

func GetVersions(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
) (*cmd.Versions, error) {
	return cmd.NewVersions(schemaRegistryClient, string(topic), string(record))
}

func GetSubjects(schemaRegistryClient srclient.ISchemaRegistryClient, topic TopicFlag) (*cmd.Subjects, error) {
	return cmd.NewSubjects(schemaRegistryClient, string(topic))
}

func GetExport(
	schemaRegistryClient srclient.ISchemaRegistryClient,
	topic TopicFlag,
	record RecordFlag,
	version VersionFlag,
) (*cmd.Export, error) {
	return cmd.NewExport(schemaRegistryClient, string(topic), string(record), int(version))
}

type App struct {
	cliApp *cli.App
	c      *dig.Container
}

func NewApp(version string) *App {
	c := dig.New()
	providers := []interface{}{
		// Flags
		GetClusterFlag,
		GetSRFlag,
		GetTopicFlag,
		GetRecordFlag,
		GetPermanentFlag,
		GetProtoFlag,
		GetVersionFlag,
		GetOutputFlag,
		GetClusterClient,
		GetSRClient,
		// Actions
		GetInspect,
		GetRegister,
		GetValidate,
		GetDelete,
		GetVersions,
		GetSubjects,
		GetExport,
	}
	for _, provider := range providers {
		c.Provide(provider)
	}

	app := &App{
		cliApp: &cli.App{
			Name:                 "Schema Registry utility",
			Version:              version,
			Usage:                "Utility for your CI/CD process to validate, register or delete Kafka protobuf schemes in the registry.",
			EnableBashCompletion: true,
		},
		c: c,
	}

	app.cliApp.Commands = []*cli.Command{
		{
			Name:      cmdRegister,
			Usage:     "Creates a subject, if one does not exists, sets a scheme for a subject or updates it.",
			ArgsUsage: "Set both the kafka cluster and the registry addresses and the protobuf file with the schema and topic&record options or set topic and record values with the separate flags.",
			Action:    makeAction(app, (*cmd.Register)(nil)),
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
			Action:    makeAction(app, (*cmd.Delete)(nil)),
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
			Action: makeAction(app, (*cmd.Validate)(nil)),
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
			Action: makeAction(app, (*cmd.Versions)(nil)),
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
			},
		},
		{
			Name:   cmdInspect,
			Usage:  "Outputs all information about the subject.",
			Action: makeAction(app, (*cmd.Inspect)(nil)),
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
			Action: makeAction(app, (*cmd.Subjects)(nil)),
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
			},
		},
		{
			Name:   cmdExport,
			Usage:  "Exports the schema value to the local file.",
			Action: makeAction(app, (*cmd.Export)(nil)),
			Flags: []cli.Flag{
				FlagSRRequired,
				FlagTopicRequired,
				FlagRecord,
				FlagVersion,
				FlagOutputRequired,
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

type Runnable interface {
	Run(context.Context) (interface{}, error)
}

func makeAction[T Runnable](a *App, _ T) cli.ActionFunc {
	return func(c *cli.Context) error {
		a.c.Provide(func() *cli.Context {
			return c
		})
		return a.c.Invoke(func(runnable T) error {
			result, err := runnable.Run(c.Context)
			if err != nil {
				return err
			}
			switch t := result.(type) {
			case string:
				log.Println(t)
			case io.Reader:
				outputFile := c.String("output")
				f, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("can not create output file: %w", err)
				}
				defer f.Close()
				if _, err := io.Copy(f, t); err != nil {
					return fmt.Errorf("can not write output to the file: %w", err)
				}
			default:
				output, err := json.MarshalIndent(result, "", "\t")
				if err != nil {
					return fmt.Errorf("can not marshall result: %w", err)
				}
				fmt.Println(string(output))
			}
			return nil
		})
	}
}
