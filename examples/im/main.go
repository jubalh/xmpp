// Copyright 2019 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// The im command sends XMPP (Jabber) messages from the command line.
// It can send instant messages to individuals and multi-user chats (MUCs),
// similar to mail(1) for SMTP (email).
//
// For more information run im -help.
package main

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/dial"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
	"mellium.im/xmpp/uri"
	"mellium.im/xmpp/version"
)

const (
	envAddr = "XMPP_ADDR"
	envPass = "XMPP_PASS"
)

type logWriter struct {
	logger *log.Logger
}

func (w logWriter) Write(p []byte) (int, error) {
	w.logger.Printf("%s", p)
	return len(p), nil
}

// messageBody is a message stanza that contains a body. It is normally used for
// chat messages.
type messageBody struct {
	stanza.Message
	Subject string `xml:"subject,omitempty"`
	Thread  string `xml:"thread,omitempty"`
	Body    string `xml:"body"`
}

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	debug := log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
	sentXML := log.New(ioutil.Discard, "SENT ", log.LstdFlags)
	recvXML := log.New(ioutil.Discard, "RECV ", log.LstdFlags)

	// Get and parse the XMPP address to send from.
	addr := os.Getenv(envAddr)
	if addr == "" {
		logger.Fatalf("environment variable $%s unset", envAddr)
	}

	parsedAddr, err := jid.Parse(addr)
	if err != nil {
		logger.Fatalf("error parsing address %q: %v", addr, err)
	}

	// Get the password to use when logging in.
	pass := os.Getenv(envPass)
	if pass == "" {
		logger.Fatalf("environment variable $%s unset", envPass)
	}

	var (
		help    bool
		rawXML  bool
		room    bool
		isURI   bool
		verbose bool
		verReq  bool
		logXML  bool
		subject string
	)
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&help, "help", help, "Show this help message")
	flags.BoolVar(&help, "h", help, "")
	flags.BoolVar(&rawXML, "xml", rawXML, "Treat the input as raw XML to be sent on the stream.")
	flags.BoolVar(&room, "room", room, "The provided JID is a multi-user chat (MUC) room.")
	flags.BoolVar(&isURI, "uri", isURI, "Parse the recipient as an XMPP URI instead of a JID.")
	flags.BoolVar(&verbose, "v", verbose, "Show verbose logging.")
	flags.BoolVar(&logXML, "vv", logXML, "Show verbose logging and sent and received XML.")
	flags.BoolVar(&verReq, "ver", verReq, "Request the software version of the remote entity instead of sending messages.")
	flags.StringVar(&addr, "addr", addr, "The XMPP address to connect to, overrides $XMPP_ADDR.")
	flags.StringVar(&subject, "subject", subject, "Set the subject of the message or chat room.")

	err = flags.Parse(os.Args[1:])
	switch err {
	case flag.ErrHelp:
		// The -h and -help flags are special cased by flags for some reason and
		// exit even if you don't register them. This should never be hit (since we
		// do register them), but handle the error just in case.
		help = true
	case nil:
	default:
		logger.Fatalf("error parsing flags: %v", err)
	}

	// If the help flag was set, just show the help message and exit.
	if help {
		printHelp(flags)
		os.Exit(0)
	}

	if verbose {
		debug.SetOutput(os.Stderr)
	}
	if logXML {
		debug.SetOutput(os.Stderr)
		sentXML.SetOutput(os.Stderr)
		recvXML.SetOutput(os.Stderr)
	}

	args := flags.Args()
	if len(args) < 1 {
		printHelp(flags)
		os.Exit(1)
	}

	var parsedToAddr, parsedAuthAddr jid.JID
	var rawMsg, thread, msgID, msgType, msgFrom string
	if isURI {
		parsedURI, err := uri.Parse(args[0])
		if err != nil {
			logger.Fatalf("error parsing %q as a URI: %v", args[0], err)
		}
		parsedToAddr = parsedURI.ToAddr
		parsedAuthAddr = parsedURI.AuthAddr
		switch parsedURI.Action {
		case "":
		case "join":
			room = true
		case "message":
			rawXML = false
			query := parsedURI.URL.Query()
			rawMsg = query.Get("body")
			subject = query.Get("subject")
			thread = query.Get("thread")
			msgID = query.Get("id")
			msgType = query.Get("type")
			msgFrom = query.Get("from")
			if msgFrom != "" {
				parsedAddr, err = jid.Parse(msgFrom)
				if err != nil {
					logger.Fatalf("error parsing %q as JID: %v", msgFrom, err)
				}
			}
		default:
			logger.Fatalf("unknown or unsupported URI action %v", parsedURI.Action)
		}
	} else {
		// Parse the recipient address as a JID.
		parsedToAddr, err = jid.Parse(args[0])
		if err != nil {
			logger.Fatalf("error parsing %q as a JID: %v", args[0], err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Login to the XMPP server.
	debug.Println("logging in…")
	dialCtx, dialCtxCancel := context.WithTimeout(ctx, 30*time.Second)
	conn, err := dial.Client(dialCtx, "tcp", parsedAddr)
	if err != nil {
		logger.Fatalf("error dialing connection: %v", err)
	}
	negotiator := xmpp.NewNegotiator(xmpp.StreamConfig{
		Features: func(*xmpp.Session, ...xmpp.StreamFeature) []xmpp.StreamFeature {
			return []xmpp.StreamFeature{
				xmpp.BindResource(),
				xmpp.StartTLS(&tls.Config{
					ServerName: parsedAddr.Domain().String(),
				}),
				xmpp.SASL(parsedAuthAddr.String(), pass, sasl.ScramSha256Plus, sasl.ScramSha1Plus, sasl.ScramSha256, sasl.ScramSha1, sasl.Plain),
			}
		},
		TeeIn:  logWriter{logger: recvXML},
		TeeOut: logWriter{logger: sentXML},
	})
	session, err := xmpp.NewSession(dialCtx, parsedAddr.Domain(), parsedAddr, conn, 0, negotiator)
	dialCtxCancel()
	if err != nil {
		logger.Fatalf("error logging in: %v", err)
	}
	go func() {
		err := session.Serve(nil)
		if err != nil {
			logger.Printf("error handling session responses: %v", err)
		}
	}()

	originJID := session.LocalAddr()

	defer func() {
		if room {
			debug.Printf("leaving the chat room %s…", addr)
			err = session.Encode(ctx, stanza.Presence{
				From: originJID,
				To:   parsedToAddr,
				Type: stanza.UnavailablePresence,
			})
			if err != nil {
				logger.Fatalf("error leaving the chat room %s: %v", addr, err)
			}
		}
		if err := session.Close(); err != nil {
			logger.Fatalf("error ending session: %v", err)
		}
		if err := session.Conn().Close(); err != nil {
			logger.Fatalf("error closing connection: %v", err)
		}
	}()

	if verReq {
		if parsedToAddr.Equal(jid.JID{}) {
			logger.Fatalf("requested software version but no address provided")
		}
		verResp, err := version.Get(ctx, session, parsedToAddr)
		if err != nil {
			logger.Fatalf("error requesting software version: %v", err)
		}
		logger.Printf("got version response:\n\tName: %s\n\tVersion: %s\n\tOS: %s", verResp.Name, verResp.Version, verResp.OS)
		return
	}

	if rawMsg == "" {
		rawMsgBuf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			logger.Fatalf("error reading message from stdin: %v", err)
		}
		rawMsg = string(rawMsgBuf)
	}
	msg := strings.ToValidUTF8(string(rawMsg), "")

	if room {
		debug.Printf("joining the chat room %s…", addr)
		// Join the MUC.
		joinPresence := struct {
			stanza.Presence
			X struct {
				History struct {
					MaxStanzas int `xml:"maxstanzas,attr"`
				} `xml:"history"`
			} `xml:"http://jabber.org/protocol/muc x"`
		}{
			Presence: stanza.Presence{
				From: originJID,
				To:   parsedToAddr,
			},
		}
		err = session.Encode(ctx, joinPresence)
		if err != nil {
			log.Fatalf("error joining MUC %s: %v", addr, err)
		}
	}

	// Send message
	if rawXML {
		err = session.Send(ctx, xml.NewDecoder(strings.NewReader(msg)))
		if err != nil {
			logger.Fatalf("error sending raw XML: %v", err)
		}
	} else {
		typ := stanza.ChatMessage
		if msgType != "" {
			typ = stanza.MessageType(msgType)
		}
		err = session.Encode(ctx, messageBody{
			Message: stanza.Message{
				ID:   msgID,
				To:   parsedToAddr,
				From: parsedAddr,
				Type: typ,
			},
			Body:    msg,
			Subject: subject,
			Thread:  thread,
		})
		if err != nil {
			logger.Fatalf("error sending message: %v", err)
		}
	}
}

func printHelp(flags *flag.FlagSet) {
	fmt.Fprintf(flags.Output(), "Usage of %s:\n", os.Args[0])
	flags.PrintDefaults()
	fmt.Printf(`
The im command sends XMPP (Jabber) messages from the command line.
It can send instant messages to individuals and multi-user chats (MUCs),
similar to mail(1) for SMTP (email).

The message will be read from stdin, and all messages will be converted to valid
UTF-8. Invalid byte sequences will be removed.

To configure the command, the following environment variables (shown with their
current value) may be set:

    XMPP_ADDR=%s
    XMPP_PASS=<not shown>
`, os.Getenv(envAddr))
}
