package converter

import "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"

// for details see: https://www.blockchain.com/explorer/api/api_websocket

func convertBlockchainBlockCreate(evt *pb.CloudEvent) {
	evt.Attributes["action"] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: "create",
		},
	}
	evt.Attributes["object"] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: "block",
		},
	}
	txt := evt.Data.(*pb.CloudEvent_TextData).TextData
	switch txt {
	case "":
		txt = "New blockchain block created\n"
	default:
		txt += "New blockchain block created\n" + txt
	}
}
