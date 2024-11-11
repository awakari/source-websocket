package converter

import (
	"errors"
	"fmt"
	"github.com/awakari/source-websocket/model"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	Convert(src string, raw map[string]any) (evt *pb.CloudEvent, err error)
}

type svc struct {
	et string
}

type ConvertFunc func(evt *pb.CloudEvent, v any) (err error)

var convSchema = map[string]any{
	"action":        toAttrStringFunc("action"),
	"best_ask":      toAttrStringFunc("bestbask"),
	"best_ask_size": toAttrStringFunc("bestasksize"),
	"best_bid":      toAttrStringFunc("bestbid"),
	"best_bid_size": toAttrStringFunc("bestbidsize"),
	"data": map[string]any{
		"properties": map[string]any{
			"auth":          toAttrStringFunc("subject"),
			"depth":         toAttrStringWithPrefixFunc("elevation", "-"),
			"flynn_region":  convertEarthquakeLocationFunc("location"),
			"lat":           toAttrStringFunc("latitude"),
			"lon":           toAttrStringFunc("longitude"),
			"mag":           convertEarthquakeMagnitudeFunc("magnitude"),
			"magtype":       toAttrStringFunc("magnitudetype"),
			"sourcecatalog": toAttrStringFunc("sourcecatalog"),
			"sourceid":      toAttrStringFunc("sourceid"),
			"time":          toAttrTimestampFunc("time"),
			"unid":          toAttrStringWithPrefixFunc("objecturl", seismicportalEuEventDetailsHtmlUnid),
		},
	},
	"high_24h":   toAttrInt32ElseStringFunc("high24h"),
	"last_size":  toAttrInt32ElseStringFunc("lastsize"),
	"low_24h":    toAttrInt32ElseStringFunc("low24h"),
	"op":         convertOpFunc("action"),
	"open_24h":   toAttrInt32ElseStringFunc("open24h"),
	"price":      convertPriceFunc("offersprice"),
	"product_id": convertTickerProductIdFunc("productid"),
	"sequence":   toAttrInt32ElseStringFunc("sequence"),
	"side":       convertTickerSideFunc("side"),
	"time":       toAttrTimestampFunc("time"),
	"trade_id":   toAttrInt32ElseStringFunc("tradeid"),
	"volume_24h": toAttrInt32ElseStringFunc("volume24h"),
	"volume_30d": toAttrInt32ElseStringFunc("volume30d"),
	"x": map[string]any{
		"txIndexes":        toAttrStringJoinedFunc("xtxindexes", " "),
		"nTx":              toAttrInt32ElseStringFunc("xntx"),
		"totalBTCSent":     toAttrInt32ElseStringFunc("xtotalbtcsent"),
		"estimatedBTCSent": toAttrInt32ElseStringFunc("xestimatedbtcsent"),
		"reward":           toStringAttrAndAppendTextLabelFunc("xreward", "Reward"),
		"size":             toAttrInt32ElseStringFunc("xsize"),
		"blockIndex":       toStringAttrAndAppendTextLabelFunc("xblockindex", "Index"),
		"prevBlockIndex":   toAttrInt32ElseStringFunc("xprevblockindex"),
		"height":           toAttrInt32ElseStringFunc("xheight"),
		"hash":             toStringAttrAndAppendTextLabelFunc("xhash", "Hash"),
		"mrklRoot":         toAttrStringFunc("xmrklroot"),
		"version":          toAttrInt32ElseStringFunc("xversion"),
		"time":             toAttrTimestampFunc("time"),
		"bits":             toAttrInt32ElseStringFunc("xbits"),
		"nonce":            toAttrInt32ElseStringFunc("nonce"),
	},
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

func convertOpFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		switch v {
		case "block":
			convertBlockchainBlockCreate(evt)
		}
		return
	}
}

func convertPriceFunc(k string) ConvertFunc {
	attrSetFunc := toAttrInt32ElseStringFunc(k)
	return func(evt *pb.CloudEvent, v any) (err error) {
		evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("Price: %s\n", v)
		err = attrSetFunc(evt, v)
		return
	}
}

func toAttrStringFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var s string
		s, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: s,
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

func toAttrStringWithPrefixFunc(k, prefix string) ConvertFunc {
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

func toAttrTimestampFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		switch vt := v.(type) {
		case int:
			if vt > 1e15 {
				// timestamp is unix micros
				vt /= 1_000_000
			}
			if vt > 1e12 {
				// timestamp is unix millis
				vt /= 1_000
			}
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: &timestamppb.Timestamp{
						Seconds: int64(vt),
					},
				},
			}
		case int32:
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: &timestamppb.Timestamp{
						Seconds: int64(vt),
					},
				},
			}
		case int64:
			if vt > 1e15 {
				// timestamp is unix micros
				vt /= 1_000_000
			}
			if vt > 1e12 {
				// timestamp is unix millis
				vt /= 1_000
			}
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: &timestamppb.Timestamp{
						Seconds: vt,
					},
				},
			}
		case float32:
			if vt > 1e15 {
				// timestamp is unix micros
				vt /= 1_000_000
			}
			if vt > 1e12 {
				// timestamp is unix millis
				vt /= 1_000
			}
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: &timestamppb.Timestamp{
						Seconds: int64(vt),
					},
				},
			}
		case float64:
			if vt > 1e15 {
				// timestamp is unix micros
				vt /= 1_000_000
			}
			if vt > 1e12 {
				// timestamp is unix millis
				vt /= 1_000
			}
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: &timestamppb.Timestamp{
						Seconds: int64(vt),
					},
				},
			}
		case string:
			var t time.Time
			t, err = time.Parse(time.RFC3339, vt)
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

func toAttrStringJoinedFunc(k, sep string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		vSlice, vSliceOk := v.([]any)
		switch vSliceOk {
		case true:
			var strs []string
			var str string
			for _, e := range vSlice {
				str, err = toString(k, e)
				if err != nil {
					break
				}
				strs = append(strs, str)
			}
			if err == nil {
				evt.Attributes[k] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: strings.Join(strs, sep),
					},
				}
			}
		default:
			err = fmt.Errorf("%w: key: %s, value %v, type: %s, expected a slice", ErrConversion, k, v, reflect.TypeOf(k))
		}
		return
	}
}

func toAttrInt32ElseStringFunc(k string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		i, ok := toInt32(v)
		switch ok {
		case true:
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeInteger{
					CeInteger: i,
				},
			}
		default:
			var s string
			s, err = toString(k, v)
			if err == nil {
				evt.Attributes[k] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: s,
					},
				}
			}
		}
		return
	}
}

func toInt32(v any) (i int32, ok bool) {
	switch vt := v.(type) {
	case bool:
		if vt {
			i = 1
		}
		ok = true
	case int:
		if vt >= math.MinInt32 && vt <= math.MaxInt32 {
			i = int32(vt)
			ok = true
		}
	case int8:
		i = int32(vt)
		ok = true
	case int16:
		i = int32(vt)
		ok = true
	case int32:
		i = vt
		ok = true
	case int64:
		if vt >= math.MinInt32 && vt <= math.MaxInt32 {
			i = int32(vt)
			ok = true
		}
	case float32:
		if vt >= math.MinInt32 && vt <= math.MaxInt32 {
			i = int32(vt)
			ok = float32(i) == vt
		}
	case float64:
		if vt >= math.MinInt32 && vt <= math.MaxInt32 {
			i = int32(vt)
			ok = float64(i) == vt
		}
	case string:
		i64, err := strconv.ParseInt(vt, 10, 32)
		if err == nil && i64 >= math.MinInt32 && i64 <= math.MaxInt32 {
			i = int32(i64)
			ok = true
		}
	}
	return
}

func toStringAttrAndAppendTextLabelFunc(k, lbl string) ConvertFunc {
	return func(evt *pb.CloudEvent, v any) (err error) {
		var s string
		s, err = toString(k, v)
		if err == nil {
			evt.Attributes[k] = &pb.CloudEventAttributeValue{
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: s,
				},
			}
			evt.Data.(*pb.CloudEvent_TextData).TextData += fmt.Sprintf("%s: %s\n", lbl, s)
		}
		return
	}
}
