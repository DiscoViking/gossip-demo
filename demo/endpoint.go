package demo

import (
	"fmt"
	"time"

	"github.com/stefankopieczek/gossip/base"
	"github.com/stefankopieczek/gossip/log"
	"github.com/stefankopieczek/gossip/transaction"
	"github.com/stefankopieczek/gossip/transport"
)


type EndPoint struct {
	// Sip Params
	DisplayName string
	UserName    string
	Host        string

	// Transport Params
	Port      uint16 // Listens on this Port.
	Transport string // Sends using this Transport. ("tcp" or "udp")

	// Internal guts
	tm       *transaction.Manager
	dialog   dialog
	dialogIx int
}

type dialog struct {
	callId    string
	to_tag    string // The tag in the To header.
	from_tag  string // The tag in the From header.
	currentTx txInfo // The current transaction.
	cseq      uint32
}

type txInfo struct {
	tx     transaction.Transaction // The underlying transaction.
	branch string                  // The via branch.
}

func (e *EndPoint) Start() error {
	trm, err := transport.NewManager(e.Transport)
	if err != nil {
		return err
	}
	tm, err := transaction.NewManager(trm, fmt.Sprintf("%v:%v", e.Host, e.Port))
	if err != nil {
		return err
	}

	e.tm = tm

	return nil
}

func (e *EndPoint) ClearDialog() {
	e.dialog = dialog{}
}

func (caller *EndPoint) Invite(callee *EndPoint) error {
	// Starting a dialog.
	callid := "thisisacall" + string(caller.dialogIx)
	tag := "tag." + caller.UserName + "." + caller.Host
	branch := "z9hG4bK.callbranch.INVITE"
	caller.dialog.callId = callid
	caller.dialog.from_tag = tag
	caller.dialog.currentTx = txInfo{}
	caller.dialog.currentTx.branch = branch

	invite := base.NewRequest(
		base.INVITE,
		&base.SipUri{
			User: base.String{S:callee.UserName},
			Host: callee.Host,
		},
		"SIP/2.0",
		[]base.SipHeader{
			Via(caller, branch),
			To(callee, caller.dialog.to_tag),
			From(caller, caller.dialog.from_tag),
			Contact(caller),
			CSeq(caller.dialog.cseq, base.INVITE),
			CallId(callid),
			ContentLength(0),
		},
		"",
	)
	caller.dialog.cseq += 1

	log.Info("Sending: %v", invite.Short())
	tx := caller.tm.Send(invite, fmt.Sprintf("%v:%v", callee.Host, callee.Port))
	caller.dialog.currentTx.tx = transaction.Transaction(tx)
	for {
		select {
		case r := <-tx.Responses():
			log.Info("Received response: %v", r.Short())
			log.Debug("Full form:\n%v\n", r.String())
			// Get To tag if present.
			tag, ok := r.Headers("To")[0].(*base.ToHeader).Params.Get("tag")
			if ok {

				switch str := tag.(type) {
				case base.String:
					caller.dialog.to_tag = str.S
//				case NoString():
//				    return str.String()
				}
			}

			switch {
			case r.StatusCode >= 300:
				// Call setup failed.
				return fmt.Errorf("callee sent negative response code %v.", r.StatusCode)
			case r.StatusCode >= 200:
				// Ack 200s manually.
				log.Info("Sending Ack")
				tx.Ack()
				return nil
			}
		case e := <-tx.Errors():
			log.Warn(e.Error())
			return e
		}
	}
}

func (caller *EndPoint) Bye(callee *EndPoint) error {
	return caller.nonInvite(callee, base.BYE)
}

func (caller *EndPoint) nonInvite(callee *EndPoint, method base.Method) error {
	caller.dialog.currentTx.branch = "z9hG4bK.callbranch." + string(method)
	request := base.NewRequest(
		method,
		&base.SipUri{
			User: base.String{S:callee.UserName},
			Host: callee.Host,
		},
		"SIP/2.0",
		[]base.SipHeader{
			Via(caller, caller.dialog.currentTx.branch),
			To(callee, caller.dialog.to_tag),
			From(caller, caller.dialog.from_tag),
			Contact(caller),
			CSeq(caller.dialog.cseq, method),
			CallId(caller.dialog.callId),
			ContentLength(0),
		},
		"",
	)
	caller.dialog.cseq += 1

	log.Info("Sending: %v", request.Short())
	tx := caller.tm.Send(request, fmt.Sprintf("%v:%v", callee.Host, callee.Port))
	caller.dialog.currentTx.tx = transaction.Transaction(tx)
	for {
		select {
		case r := <-tx.Responses():
			log.Info("Received response: %v", r.Short())
			log.Debug("Full form:\n%v\n", r.String())
			switch {
			case r.StatusCode >= 300:
				// Failure (or redirect).
				return fmt.Errorf("callee sent negative response code %v.", r.StatusCode)
			case r.StatusCode >= 200:
				// Success.
				log.Info("Successful transaction")
				return nil
			}
		case e := <-tx.Errors():
			log.Warn(e.Error())
			return e
		}
	}
}

// Server side function.

func (e *EndPoint) ServeInvite() {
	log.Info("Listening for incoming requests...")
	tx := <-e.tm.Requests()
	r := tx.Origin()
	log.Info("Received request: %v", r.Short())
	log.Debug("Full form:\n%v\n", r.String())

	e.dialog.callId = string(*r.Headers("Call-Id")[0].(*base.CallId))

	// Send a 200 OK
	resp := base.NewResponse(
		"SIP/2.0",
		200,
		"OK",
		[]base.SipHeader{},
		"",
	)

	base.CopyHeaders("Via", tx.Origin(), resp)
	base.CopyHeaders("From", tx.Origin(), resp)
	base.CopyHeaders("To", tx.Origin(), resp)
	base.CopyHeaders("Call-Id", tx.Origin(), resp)
	base.CopyHeaders("CSeq", tx.Origin(), resp)
	resp.AddHeader(
		&base.ContactHeader{
			DisplayName: base.String{S:e.DisplayName},
			Address: &base.SipUri{
				User: base.String{S:e.UserName},
				Host: e.Host,
			},
		},
	)

	log.Info("Sending 200 OK")
	<-time.After(1 * time.Second)
	tx.Respond(resp)

	ack := <-tx.Ack()

	log.Info("Received ACK")
	log.Debug("Full form:\n%v\n", ack.String())
}

func (e *EndPoint) ServeNonInvite() {
	log.Info("Listening for incoming requests...")
	tx := <-e.tm.Requests()
	r := tx.Origin()
	log.Info("Received request: %v", r.Short())
	log.Debug("Full form:\n%v\n", r.String())

	// Send a 200 OK
	resp := base.NewResponse(
		"SIP/2.0",
		200,
		"OK",
		[]base.SipHeader{},
		"",
	)

	base.CopyHeaders("Via", tx.Origin(), resp)
	base.CopyHeaders("From", tx.Origin(), resp)
	base.CopyHeaders("To", tx.Origin(), resp)
	base.CopyHeaders("Call-Id", tx.Origin(), resp)
	base.CopyHeaders("CSeq", tx.Origin(), resp)
	resp.AddHeader(
		&base.ContactHeader{
			DisplayName: base.String{S:e.DisplayName},
			Address: &base.SipUri{
				User: base.String{S:e.UserName},
				Host: e.Host,
			},
		},
	)

	log.Info("Sending 200 OK")
	<-time.After(1 * time.Second)
	tx.Respond(resp)
}

