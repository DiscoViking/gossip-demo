package demo

import "github.com/stefankopieczek/gossip/base"


// Utility methods for creating headers.

func Via(e *EndPoint, branch string) *base.ViaHeader {
	return &base.ViaHeader{
		&base.ViaHop{
			ProtocolName:    "SIP",
			ProtocolVersion: "2.0",
			Transport:       e.Transport,
			Host:            e.Host,
			Port:            &e.Port,
			Params: base.NewParams().Add("branch",base.String{S:branch}),
		},
	}
}

func To(e *EndPoint, tag string) *base.ToHeader {
	header := &base.ToHeader{
		DisplayName: base.String{S:e.DisplayName},
		Address: &base.SipUri{
			User:      base.String{S:e.UserName},
			Host:      e.Host,
			UriParams: base.NewParams(),
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		header.Params.Add("tag",base.String{S:tag})
	}

	return header
}

func From(e *EndPoint, tag string) *base.FromHeader {
	header := &base.FromHeader{
		DisplayName: base.String{S:e.DisplayName},
		Address: &base.SipUri{
			User: base.String{S:e.UserName},
			Host: e.Host,
			UriParams: base.NewParams().Add("Transport",base.String{S:e.Transport}),
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		header.Params.Add("tag",base.String{S:tag})
	}

	return header
}

func Contact(e *EndPoint) *base.ContactHeader {
	return &base.ContactHeader{
		DisplayName: base.String{S:e.DisplayName},
		Address: &base.SipUri{
			User: base.String{S:e.UserName},
			Host: e.Host,
		},
	}
}

func CSeq(seqno uint32, method base.Method) *base.CSeq {
	return &base.CSeq{
		SeqNo:      seqno,
		MethodName: method,
	}
}

func CallId(callid string) *base.CallId {
	header := base.CallId(callid)
	return &header
}

func ContentLength(l uint32) base.ContentLength {
	return base.ContentLength(l)
}

