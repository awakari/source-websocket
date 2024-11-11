package converter

import (
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

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
	evt.Data.(*pb.CloudEvent_TextData).TextData = "New blockchain block created\n" + evt.Data.(*pb.CloudEvent_TextData).TextData
}
