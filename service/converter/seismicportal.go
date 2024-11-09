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
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Earthquake location: %s\n", l)
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
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Earthquake magnitude: %s\n", m)
		}
		return
	}
}
