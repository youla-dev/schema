// Package protoschema implements helpers to access the topic&record option values
// from the Protocol Buffers original scheme or from the generated golang code.

package protoschema

import (
	"bytes"
	"context"
	"fmt"

	eproto "github.com/emicklei/proto"
	"github.com/golang/protobuf/descriptor"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Parse loads topic&record option values from the original proto file.
func Parse(ctx context.Context, protobuf []byte) (string, string, error) {
	parser := eproto.NewParser(bytes.NewBuffer(protobuf))
	definition, err := parser.Parse()

	if err != nil {
		return "", "", fmt.Errorf("can not parse: %w", err)
	}

	var topic, record string

	eproto.Walk(definition,
		eproto.WithOption(func(option *eproto.Option) {
			if option.Name == "(topic)" {
				topic = option.Constant.Source
			}
			if option.Name == "(record)" {
				record = option.Constant.Source
			}
		}),
	)

	return topic, record, nil
}

// ExtractTopicRecord loads topic&record option values from the generated golang code.
// If the recordOption is not specified, the second output string would be empty.
func ExtractTopicRecord(m interface{}, topicOption, recordOption protoreflect.ExtensionType) (string, string, error) {
	_, md := descriptor.MessageDescriptorProto(m)
	var topic, record string
	topicRaw := proto.GetExtension(md.GetOptions(), topicOption)
	if topicRaw != nil {
		topic = topicRaw.(string)
	}
	if recordOption != nil {
		recordRaw := proto.GetExtension(md.GetOptions(), recordOption)
		if recordRaw != nil {
			record = recordRaw.(string)
		}
	}

	return topic, record, nil
}
