package main

import "github.com/stefankopieczek/gossip/base"

// Utility methods for creating headers.

func Via(e *endpoint, branch string) *base.ViaHeader {
	params := base.NewParams()
	items := params.Items()
	items["branch"] = &base.String{S: branch}
	return &base.ViaHeader{
		&base.ViaHop{
			ProtocolName:    "SIP",
			ProtocolVersion: "2.0",
			Transport:       e.transport,
			Host:            e.host,
			Port:            &e.port,
			Params:          params,
		},
	}
}

func To(e *endpoint, tag string) *base.ToHeader {
	params := base.NewParams()
	items := params.Items()
	items["transport"] = &base.String{S: e.transport}
	header := &base.ToHeader{
		DisplayName: &base.String{S: e.displayName},
		Address: &base.SipUri{
			User:      &base.String{S: e.username},
			Host:      e.host,
			UriParams: params,
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		items := header.Params.Items()
		items["tag"] = &base.String{S: tag}
	}

	return header
}

func From(e *endpoint, tag string) *base.FromHeader {
	params := base.NewParams()
	items := params.Items()
	items["transport"] = &base.String{S: e.transport}
	header := &base.FromHeader{
		DisplayName: &base.String{S: e.displayName},
		Address: &base.SipUri{
			User: &base.String{S: e.username},
			Host: e.host,
			UriParams: params,
		},
		Params: base.NewParams(),
	}

	if tag != "" {
		items := header.Params.Items()
		items["tag"] = &base.String{S: tag}
	}

	return header
}

func Contact(e *endpoint) *base.ContactHeader {
	return &base.ContactHeader{
		DisplayName: &base.String{S: e.displayName},
		Address: &base.SipUri{
			User: &base.String{S: e.username},
			Host: e.host,
			Port: &e.port,
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
