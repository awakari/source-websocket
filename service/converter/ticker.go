package converter

import (
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

const typeTicker = "ticker"

func convertTickerType(evt *pb.CloudEvent, k string) {
	evt.Attributes[k] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: typeTicker,
		},
	}
}

func convertTickerProductIdFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var pid string
		pid, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: pid,
				},
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Ticker product id: %s\n", pid)
		}
		return
	}
}

func convertTickerSideFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var pid string
		pid, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: pid,
				},
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Ticker side: %s\n", pid)
		}
		return
	}
}
