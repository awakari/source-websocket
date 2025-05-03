package converter

import (
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

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
			txt := evt.Data.(*pb.CloudEvent_TextData).TextData
			switch txt {
			case "":
				txt = fmt.Sprintf("Ticker\nProduct id: %s\n", pid)
			default:
				txt += fmt.Sprintf("Product id: %s\n", pid)
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData = txt
		}
		return
	}
}

func convertTickerSideFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var side string
		side, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: side,
				},
			}
			txt := evt.Data.(*pb.CloudEvent_TextData).TextData
			switch txt {
			case "":
				txt = fmt.Sprintf("Ticker\nSide: %s\n", side)
			default:
				txt += fmt.Sprintf("Side: %s\n", side)
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData = txt
		}
		return
	}
}
