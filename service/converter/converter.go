package converter

import (
	"errors"
	"fmt"
	"github.com/awakari/source-websocket/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strconv"
	"time"
)

type Service interface {
	Convert(src string, raw map[string]any) (evt *pb.CloudEvent, err error)
}

type svc struct {
	et string
}

type ConvertFunc func(evt *pb.CloudEvent, v any) (err error)

const seismicportalEuEventDetailsHtmlUnid = "https://www.seismicportal.eu/eventdetails.html?unid="
const typeTicker = "ticker"

var convSchema = map[string]any{
	"action":        toStringFunc("action"),
	"best_ask":      toStringFunc("bestbask"),
	"best_ask_size": toStringFunc("bestasksize"),
	"best_bid":      toStringFunc("bestbid"),
	"best_bid_size": toStringFunc("bestbidsize"),
	"data": map[string]any{
		"properties": map[string]any{
			"auth":          toStringFunc("subject"),
			"depth":         toStringWithPrefixFunc("elevation", "-"),
			"flynn_region":  convertEarthquakeLocationFunc("location"),
			"lat":           toStringFunc("latitude"),
			"lon":           toStringFunc("longitude"),
			"mag":           convertEarthquakeMagnitudeFunc("magnitude"),
			"magtype":       toStringFunc("magnitudetype"),
			"sourcecatalog": toStringFunc("sourcecatalog"),
			"sourceid":      toStringFunc("sourceid"),
			"time":          toTimestampFunc("time"),
			"unid":          toStringWithPrefixFunc("objecturl", seismicportalEuEventDetailsHtmlUnid),
		},
	},
	"high_24h":   toStringFunc("high24h"),
	"last_size":  toStringFunc("lastsize"),
	"low_24h":    toStringFunc("low24h"),
	"open_24h":   toStringFunc("open24h"),
	"price":      convertPrice("offersprice"),
	"product_id": convertTickerProductIdFunc("productid"),
	"sequence":   toStringFunc("sequence"),
	"side":       convertTickerSideFunc("side"),
	"time":       toTimestampFunc("time"),
	"trade_id":   toStringFunc("tradeid"),
	"type":       convertTypeFunc("type"),
	"volume_24h": toStringFunc("volume24h"),
	"volume_30d": toStringFunc("volume30d"),
}

var ErrConversion = errors.New("conversion failure")

func NewService(et string) Service {
	return svc{
		et: et,
	}
}

func (s svc) Convert(src string, raw map[string]any) (evt *pb.CloudEvent, err error) {
	evt = &pb.CloudEvent{
		Id:          ksuid.New().String(),
		Source:      src,
		SpecVersion: model.CeSpecVersion,
		Type:        s.et,
		Attributes:  make(map[string]*pb.CloudEventAttributeValue),
		Data:        &pb.CloudEvent_TextData{},
	}
	err = convert(evt, raw, convSchema)
	return
}

func convert(evt *pb.CloudEvent, node map[string]any, schema map[string]any) (err error) {
	for k, v := range node {
		schemaChild, schemaChildOk := schema[k]
		if schemaChildOk {
			switch schemaChildT := schemaChild.(type) {
			case ConvertFunc:
				err = errors.Join(err, schemaChildT(evt, v))
			case map[string]any:
				branch, branchOk := v.(map[string]any)
				if branchOk {
					err = errors.Join(convert(evt, branch, schemaChildT))
				}
			}
		}
	}
	return
}

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
	return func(evt *pb.CloudEvent, v any) (err error) {
		var m string
		m, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: m,
				},
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Earthquake magnitude: %s\n", m)
		}
		return
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

func convertPrice(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var p string
		p, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: p,
				},
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Price: %s\n", p)
		}
		return
	}
}

func convertTypeFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		switch vt := v.(type) {
		case string:
			if vt == typeTicker {
				evt.Attributes[k] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: typeTicker,
					},
				}
			}
		}
		return
	}
}

func toStringFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var str string
		str, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: str,
				},
			}
		}
		return
	}
}

func toString(k string, v any) (str string, err error) {
	switch vt := v.(type) {
	case bool:
		str = strconv.FormatBool(vt)
	case int:
		str = strconv.Itoa(vt)
	case int8:
		str = strconv.Itoa(int(vt))
	case int16:
		str = strconv.Itoa(int(vt))
	case int32:
		str = strconv.Itoa(int(vt))
	case int64:
		str = strconv.FormatInt(vt, 10)
	case float32:
		switch float32(int(vt)) == vt {
		case true:
			str = strconv.Itoa(int(vt))
		default:
			str = fmt.Sprintf("%f", vt)
		}
	case float64:
		switch float64(int(vt)) == vt {
		case true:
			str = strconv.Itoa(int(vt))
		default:
			str = fmt.Sprintf("%f", vt)
		}
	case string:
		str = vt
	default:
		err = fmt.Errorf("%w: key: %s, value: %v, type: %s, expected: string/bool/int/float", ErrConversion, k, v, reflect.TypeOf(v))
	}
	return
}

func toStringWithPrefixFunc(k, prefix string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var str string
		str, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: prefix + str,
				},
			}
		}
		return
	}
}

func toTimestampFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		str, strOk := v.(string)
		switch strOk {
		case true:
			var t time.Time
			t, err = time.Parse(time.RFC3339, str)
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(t),
				},
			}
		default:
			err = fmt.Errorf("%w: key: %s, value %v, type: %s, expected timestamp in RFC3339 format", ErrConversion, k, v, reflect.TypeOf(k))
		}
		return
	}
}
