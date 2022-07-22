# Schema

[English](README.md)

Утилита для работы с Confluent Schema Registry .

> В Schema Registry основной сущностью являются `subject`, который связан с топиком. Схемы для `key` и `value` сообщения связаны в свою очередь с subject.
> В этой утилите используется [TopicRecordNameStrategy](https://docs.confluent.io/platform/current/schema-registry/serdes-develop/index.html#overview),
> которая подразумевает, что subject для value формируется `<topic name>-<fully-qualified record name>-value`.
> В этой утилите вместо subject везде фигурируют Kafka topic и record, а все схемы создаются для value сообщения.
> Создать схему для key или с другой стратегией именования пока нельзя.

# Функции

- `validate` - проверяет, что схема совместима с существующей
- `register` - регистрация схемы в SR
- `subjects` - получить список схем для топика
- `versions` - получить список версий для схемы
- `inspect` - информация о схеме
- `export` - экспортировать схему из SR в локальный файл

Конфигурация через cli параметры или через переменные окружения.

## Validate

1. Загружает protobuf (`--proto`).
2. Проверяет, что топик существует. Если не существует -- считается, что схема валидна.
3. Проверяет, что схема уже существует в SR. Если не существует -- считается, что схема валидна.
4. Проверяет с помощью SR, совместима ли новая схема с существующей.

Пример `schema validate --proto message.proto --topic current_weather --cluster localhost:9092 --sr http://localhost:8081`.

Если SR отвечает, что обновлённая схема не совместима, хотя должна быть, возможно, необходимо проверить Compatibility level и установить нужный:

```bash
curl -X PUT -H "Content-Type: application/vnd.schemaregistry.v1+json" --data '{"compatibility": "FORWARD"}' http://localhost:8081/config
```

## Register

Выполняет те же, шаги, что и validate + регистрирует в конце новую версию схемы для топика.

Пример `schema register --proto schema.proto --topic current_weather --cluster localhost:9092 --sr http://localhost:8081`.

## Subjects

Выводит список существующих схем, релевантных топику.

Пример `schema subjects --topic current_weather --sr http://localhost:8081`.

## Versions

Выводит список существующих версий для схемы.

Пример `schema versions --topic current_weather --sr http://localhost:8081`.

## Inspect

Выводит информацию о схеме: список версий, id и саму схему. Кроме топика можно выбрать желаемую версию схему, по-умолчанию `latest`.

Пример `schema inspect --topic current_weather --sr http://localhost:8081 --version=1`.

## Export

Экспортирует схему для топика в файл. Версия указывается отдельным параметром и по-умолчанию `latest`.
Можно использовать для автоматизации вместе с `protoc`.

Пример `schema export --topic current_weather --sr http://localhost:8081 --version=1 --output message.proto`.

# Определение имени топика и записи (record)

Топик и запись могут быть переданы в schema через аргументы, переменные окружения или определены в proto файле.
Общие для всех команд параметры:

- Топик: `--topic` или переменная `TOPIC`.
- Запись: `--record` или переменная `RECORD`.

## Определение в proto файле.

Для улучшения автоматизации в CI/CD при работе с несколькими записями в одном топике, название топика и записи можно
указать с помощью блока `option` в proto файле. Подробности работы с опциями в [документации Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3#customoptions).

Пример определения названий:

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

Если не передать конкретный топик и запись через аргументы, утилита будет искать опции `topic` и `record` в `message` и использовать их.

Если используется несколько схем для одного сервиса, их необходимо разметить в отдельных файлах, а опции определить один раз:

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

Далее подключить файл с сигнатурами опций где нужно с помощью `import "topic_option.proto";`.

> Пока нет возможности указать несколько message в одном файле, см. секцию [TODO](#todo).

### Как получить значение опции в коде

Protobuf файл с опциями генерируется стандартной командой protoc. Далее в Golang значения опций `topic` и `record` можно получить следующим образом:

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

где `Currency` -- модель сгенерированная из предыдущего примера. 

Либо можно воспользоваться функцией `lib/protoschema.ExtractTopicRecord` из этого проекта.

# Запуск

## Конфигурация

Утилита используется как с cli флагами (`--sr http://localhost:8081`), так и с переменными окружения (`SCHEMA_REGISTRY=http://localhost:8081`).

Переменные окружения `SCHEMA_REGISTRY` и `CLUSTER` можно определить в dotenv файле, например `.schema_config`,
путь к которому передать через переменную `SCHEMA_CONFIG=/home/user/.schema_config`.

## Docker

Запуск с помощью docker:

```bash
docker run youlatech/schema:latest schema
```

## Локально

```bash
go build -ldflags="-X 'main.Version=0.5'" -o schema ./cmd/schema/*.go
```

## В CD/CD

Любой параметр в `schema` можно передать как аргументом (e.g. `--cluster localhost:9092`), так и через переменную окружения, например `CLUSTER=localhost:9092`.
Например, запуск валидации будет выглядеть так:

```bash
PROTO=message.proto TOPIC=current_weather CLUSTER=localhost:9092 SCHEMA_REGISTRY=http://localhost:8081 schema validate
```

Пример для Gitlab CI:

```yaml
producer1:schema-validate:
  image: youlatech/schema:latest
  script:
    - schema validate --proto services/$PROTO --cluster ${KAFKA_CLUSTER} --sr ${SR}
  variables:
    PROTO: producer1/cmd/producer1/weather_message.proto
```

# Тестирование со schema-registry.

Если тестовый schema registry не работает, его легко развернуть вместе с kafka локально через [docker-compose](https://docs.confluent.io/platform/current/quickstart/ce-docker-quickstart.html).
