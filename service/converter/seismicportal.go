package converter

import (
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

const seismicportalEuEventDetailsHtmlUnid = "https://www.seismicportal.eu/eventdetails.html?unid="

func convertEarthquakeLocationFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var l string
		l, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: l,
				},
			}
			txt := evt.Data.(*pb.CloudEvent_TextData).TextData
			switch txt {
			case "":
				txt = fmt.Sprintf("Earthquake\nLocation: %s\n", l)
			default:
				txt += fmt.Sprintf("Location: %s\n", l)
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData = txt
		}
		return
	}
}

func convertEarthquakeMagnitudeFunc(k string) ConvertFunc {
	attrSetFunc := toInt32ElseStringFunc(k)
	return func(evt *pb.CloudEvent, v any) (err error) {
		var m string
		m, err = toString(k, v)
		if err == nil {
			err = attrSetFunc(evt, v)
			txt := evt.Data.(*pb.CloudEvent_TextData).TextData
			switch txt {
			case "":
				txt = fmt.Sprintf("Earthquake\nMagnitude: %s\n", m)
			default:
				txt += fmt.Sprintf("Magnitude: %s\n", m)
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData = txt
		}
		return
	}
}
