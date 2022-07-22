# Schema

[Русский](README_ru.md)

Utility for your CI/CD process to validate, register or delete Kafka protobuf schemes in the registry.

> Schema registry mainly operates with the "subject" concept related to a topic. Therefore, key and value schemes are related to the subject.
> This utility uses [TopicRecordNameStrategy](https://docs.confluent.io/platform/current/schema-registry/serdes-develop/index.html#overview), 
> which means that the subject for the value is defined as `<topic name>-<fully-qualified record name>-value`.
> This utility operates only with topic and record values and hides subject name generation. 
> Also, the ability to create schemes for the topic keys is not realized for now.

# Commands

- `validate` - Validates the topic to exist and the schema changes compatibility with existing version.
- `register` - Creates a subject, if one does not exist, sets a scheme for a subject or updates it.
- `subjects` - Lists available subjects for the topic.
- `versions` - Lists available versions for the subject.
- `inspect` - Outputs all information about the subject.
- `export` - Exports the schema value to the local file.

The application is being configured via CLI flags, environment variables or dotenv file.

## Validate

1. Loads protobuf file (`--proto`).
2. Validates for the topic existence. The schema is considered to be valid, if the topic does not exist.
3. Validates for the subject existence. The schema is considered to be valid, if the subject does not exist in Schema Registry.
4. Calls Schema Registry to verify the compatibility of the new version of the schema.

Example `schema validate --proto message.proto --topic current_weather --cluster localhost:9092 --sr http://localhost:8081`.


Try to check Compatibility level and fix it if updated schema is not compatible with the previous one despite of mistakes absence:

```bash
curl -X PUT -H "Content-Type: application/vnd.schemaregistry.v1+json" --data '{"compatibility": "FORWARD"}' http://localhost:8081/config
```

## Register

Performs the same steps as "validate" command and registers the scheme or new version in the end.

Example `schema register --proto schema.proto --topic current_weather --cluster localhost:9092 --sr http://localhost:8081`.

## Subjects

Lists available subjects for the topic.

Example `schema subjects --topic current_weather --sr http://localhost:8081`.

## Versions

Lists available versions for the subject.

Example `schema versions --topic current_weather --sr http://localhost:8081`.

## Inspect

Outputs all information about the subject: versions list, id and the scheme value. The version by default is `latest`.

Example `schema inspect --topic current_weather --sr http://localhost:8081 --version=1`.

## Export

Exports the schema value to the local file. Schema version is to be set with `--version` flag or`latest` by default.
The command can be used for the automation of code generation on combination with `protoc` utility.

Example `schema export --topic current_weather --sr http://localhost:8081 --version=1 --output message.proto`.

# How to set topic and record

Topic&record can be set with CLI flags, environmental variables or with proto file.

The common flags:
- Topic: flag `--topic` or variable `TOPIC`.
- Record: flag`--record` or variable `RECORD`.

## Topic&record definition with proto file

For CI/CD automation you can set the separate topic&record values for the each schema with `option` field in the proto file.
More information about Protobuf options can be found in the [Protocol Buffers documentation](https://developers.google.com/protocol-buffers/docs/proto3#customoptions).

Example:

```protobuf
syntax = "proto3";

import "google/protobuf/descriptor.proto";

extend google.protobuf.MessageOptions {
  string topic = 50001;
  string record = 50002;
}

message Currency {
  option (topic) = "weather";
  option (record) = "currency";

  string value = 1;
}
```

The utility would use Protobuf options if the flags `--topic` and `--recourd` are not present.

If your service uses several schemes for the one topic you need to place the to the separate proto files.
The options must be defined single time as in the following example.

```protobuf
syntax = "proto3";

package producer1;

option go_package = "main";

import "google/protobuf/descriptor.proto";

extend google.protobuf.MessageOptions {
  string topic = 50001;
  string record = 50002;
}
```

The options signatures would be loaded with`import "topic_option.proto";` in every message definition file.

### How to get options values in the code

1. Message code must be generated with the standard protoc utility.
2. The topic&record option values can be accessed with the following code.

```go
package main

import (
	"google.golang.org/protobuf/proto"
	"github.com/golang/protobuf/descriptor"
)

func main() {
	getTopicRecord((*Currency)(nil))
}

func getTopicRecord(m interface{}) (string, string) {
	_, md := descriptor.MessageDescriptorProto(m)
	record := proto.GetExtension(md.GetOptions(), E_Record)
	topic := proto.GetExtension(md.GetOptions(), E_Topic)
	return topic.(string), record.(string)
}
```

Here`Currency` is the model generated within the previous section.

Another way to get values is to use `lib/protoschema.ExtractTopicRecord` function.

# Run

## Configuration

Utility gets configuration values with CLI flags (e.g. `--sr http://localhost:8081`) or with environmental variables (e.g. `SCHEMA_REGISTRY=http://localhost:8081`).

The common environmental variables `SCHEMA_REGISTRY` and `CLUSTER` can be placed to the dotenv file located with `SCHEMA_CONFIG=/home/user/.schema_config` env variable.

## Docker

Run with docker:

```bash
docker run youlatech/schema:latest schema
```

## Local build

```bash
go build -ldflags="-X 'main.Version=0.5'" -o schema ./cmd/schema/*.go
```

The last binary can also be downloaded from [the releases page](https://github.com/youla-dev/schema/releases).

## CD/CD

Example for the GitLab CI:

```yaml
producer1:schema-validate:
  image: youlatech/schema:latest
  script:
    - schema validate --proto services/$PROTO --cluster ${KAFKA_CLUSTER} --sr ${SR}
  variables:
    PROTO: producer1/cmd/producer1/weather_message.proto
```

## How to run Confluent Schema Registry.

You can try to set up local Confluent Schema registry environment with the Confluent [docker-compose](https://docs.confluent.io/platform/current/quickstart/ce-docker-quickstart.html) example.