# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Breaking

- color: change list of color vision deficiencies from uint8 to a new type


### Added

- delay: new package implementing [XEP-0203: Delayed Delivery]
- disco: new package implementing [XEP-0030: Service Discovery]
- paging: new package implementing [XEP-0059: Result Set Management]
- stanza: new functions `AddID` and `AddOriginID` to support unique and stable
  stanza IDs
- stanza: ability to compare errors with `errors.Is`
- styling: satisfy `fmt.Stringer` for the `Style` type
- version: new package implementing [XEP-0092: Software Version]
- xmpp: satisfy `fmt.Stringer` for the `SessionState` type
- xmpp: new `UnmarshalIQ`, `UnmarshalIQElement`, `IterIQ`, and `IterIQElement`
  methods
- xtime: times can now be marshaled and unmarshaled as XML attributes


### Fixed

- form: if no field type is set the correct default (text-single) is used
- xmpp: unknown IQ error responses are now sent to the correct address


[XEP-0030: Service Discovery]: https://xmpp.org/extensions/xep-0030.html
[XEP-0059: Result Set Management]: https://xmpp.org/extensions/xep-0059.html
[XEP-0092: Software Version]: https://xmpp.org/extensions/xep-0092.html
[XEP-0203: Delayed Delivery]: https://xmpp.org/extensions/xep-0203.html


## v0.18.0 — 2021-02-14

### Breaking

- component: `NewClientSession` has been split in to `NewSession` and
  `ReceiveSession`
- mux: fixed the signature of the `IQFunc` option to differentiate it from the
  normal `IQ` option
- stream: the `SeeOtherHost` function signature has changed now that payload is
  available for all stream errors
- stream: the `TokenReader` method signature has changed to make it match the
  `xmlstream.Marshaler` interface
- stream: renamed `ErrorNS` to `NSError` for consistency
- xmpp: the `Negotiator` type signature has changed
- xmpp: the `StreamConfig.Features` field is now a callback instead of a slice
- xmpp: the `StreamFeature` type's `Parse` function now takes an xml.Decoder
  instead of an `xml.TokenReader`
- xmpp: IQs sent during stream negotiation are now matched against registered
  stream features based on the namespace of their payload
- xmpp: the session negotiation functions have been renamed and revamped, see
  the main package docs for a comprehensive overview
- xtime: change the `Time` type to fix unmarshaling


### Added

- internal/integration/mcabber: [Mcabber] support for integration tests
- oob: implementations of `xmlstream.Marshaler` and `xmlstream.WriterTo` for the
  types `IQ`, `Query`, and Data
- receipts: add `Request` and `Requested` to add receipt requests to messages
  without waiting on a response
- roster: add `Set` and `Delete` functions for roster management
- roster: support multiple groups
- s2s: added a new package implementing Bidi
- stanza: add `Error` method on the `IQ` type
- stream: new `InnerXML` and `ApplicationError` methods on `Error` provide a way
  to easily construct customized stream errors
- stream: ability to compare errors with `errors.Is`
- stream: support adding human-readable text to errors
- stream: add `Info` block for use by custom `Negotiators`
- styling: add `Disable` and `Unstyled` to support disabling styling on some
  messages
- websocket: added a new package for dialing WebSocket connections
- xmpp: `SetReadDeadline` and `SetWriteDeadline` are now proxied even if the
  underlying connection is not a `net.Conn`
- xmpp: all sent stanzas are now given randomly generated IDs if no ID was
  provided (not just IQs)
- xmpp: the start token that triggers a call to `Negotiate` is now no longer
  popped from the stream before the actual call to `Negotiate` for server side
  stream features
- xmpp: a `SASLServer` stream feature for authenticating users
- xmpp: two methods for getting the stream IDs, `InSID` and `OutSID`
- xmpp: a new state bit, `s2s` to indicate whether the session is a
  server-to-server connection

[Mcabber]: https://mcabber.com/


### Changed

- dial: resolving SRV records for XMPP over implicit or opportunistic TLS is
  now done concurrently to make the initial connection faster
- dial: the fallback for dialing an XMPP connection when no SRV records exist is
  now more robust
- xmpp: stanzas sent over S2S connections now always have the "from" attribute
  set
- xmpp: the default negotiator now supports negotiating the WebSocket
  subprotocol when the `WebSocket` option is set


### Fixed

- component: the `NewSession` function (previously `NewClientSession`) now
  correctly marks the connection as having been received or initiated
- component: the wrong error is no longer returned if Prosody sends an error
  immediately after the start of a stream with no ID
- dial: dialing an XMPP connection where no xmpps SRV records exist no longer
  results in an error (fallback works correctly)
- roster: fix infinite loop marshaling lists of items
- stream: errors are now unmarshaled correctly
- xmpp: the Encode methods no longer sometimes duplicate the xmlns attribute
- xmpp: stream errors are now unmarshaled and returned from `Serve` and during
  session negotiation
- xmpp: XML tokens written directly to the session are now always flushed to the
  underlying connection when the token writer is closed
- xmpp: stream feature parse functions can no longer read beyond the end of the
  feature element
- xmpp: the resource binding feature was previously broken for server sessions.
- xmpp: the "to" and "from" attributes on incoming streams are now verified and
  an error is returned if they change between stream restarts
- xmpp: empty "from" attributes are now removed before marshaling so that a
  default value can be set (if applicable)


## v0.17.1 — 2020-11-21

### Breaking

- roster: remove workaround for a bug in Go versions prior to 1.14 which is now
  the earliest supported version
- xmpp: the `Encode` and `EncodeElement` methods now take a context and respect
  its deadline

### Added

- internal/integration: new package for writing integration tests
- internal/integration/ejabberd: [Ejabberd] support for integration tests
- internal/integration/prosody: [Prosŏdy] support for integration tests
- internal/xmpptest: new `ClientServer` for testing two connected sessions
- xmpp: new `EncodeIQ` and `EncodeIQElement` methods

[Ejabberd]: https://www.ejabberd.im/
[Prosŏdy]: https://prosody.im/


### Fixed

- stanza: converting stanzas with empty to/from attributes no longer fails
- xmpp: fixed data race that could result in invalid session state and lead to
  writes on a closed session and other state related issues
- xmpp: the `Send` and `SendElement` methods now respect the context deadline


## v0.17.0 — 2020-11-11

### Breaking

- sasl2: removed experimental package
- xmpp: removed option to make STARTTLS feature optional


### Added

- styling: new package implementing [XEP-0393: Message Styling]
- xmpp: `ConnectionState` method


### Fixed

- roster: iters that have errored no longer panic when closed
- xmpp: using TeeIn/TeeOut no longer breaks SCRAM based SASL mechanisms
- xmpp: stream negotiation no longer fails when the only required features
  cannot yet be negotiated because they depend on optional features


[XEP-0393: Message Styling]: https://xmpp.org/extensions/xep-0393.html


## v0.16.0 — 2020-03-08

### Breaking

- xmpp: the end element is now included in the token stream passed to handlers
- xmpp: SendElement now wraps the entire stream, not just the first element


### Added

- receipts: new package implementing [XEP-0333: Chat Markers]
- roster: add handler and mux option for roster pushes


[XEP-0333: Chat Markers]: https://xmpp.org/extensions/xep-0333.html


### Fixed

- mux: fix broken `Decode` and possible infinite loop due to cutting off the
  last token in a buffered XML token stream
- roster: work around a bug in Go 1.13 where `io.EOF` may be returned from the
  XML decoder


## v0.15.0 — 2020-02-28

### Breaking

- all: dropped support for versions of Go before 1.13
- mux: move `Wrap{IQ,Presence,Message}` functions to methods on the stanza types


### Added

- mux: ability to select handlers by stanza payload
- mux: new handler types and API
- ping: a function for easily encoding pings and handling errors
- ping: a handler and mux option for responding to pings
- stanza: ability to convert stanzas to/from `xml.StartElement`
- stanza: API to simplify replying to IQs
- uri: new package for parsing XMPP URI's and IRI's
- xtime: new package for handling [XEP-0202: Entity Time] and [XEP-0082: XMPP Date and Time Profiles]


[XEP-0202: Entity Time]: https://xmpp.org/extensions/xep-0202.html
[XEP-0082: XMPP Date and Time Profiles]: https://xmpp.org/extensions/xep-0082.html


### Fixed

- dial: if a port number is present in a JID it was previously ignored


## v0.14.0 — 2019-08-18

### Breaking

- ping: remove `IQ` function and replace with struct based API


### Added

- ping: add `IQ` struct based encoding API


### Changed

- stanza: a zero value `IQType` now marshals as "get"
- xmpp: read timeouts are now returned instead of ignored


### Fixed

- dial: fix broken fallback to domainpart
- xmpp: allow whitespace keepalives
- roster: the iterator now correctly closes the underlying TokenReadCloser
- xmpp: fix bug where stream processing could stop after an IQ was received


## v0.13.0 — 2019-07-27

### Breaking

- xmpp: change `Handler` to take an `xmlstream.TokenReadEncoder`
- xmpp: replace `EncodeToken` and `Flush` with `TokenWriter`
- xmpp: replace `Token` with `TokenReader`


### Added

- examples/echobot: add graceful shutdown on SIGINT
- xmpp: `Encode` and `EncodeElement` methods


### Changed

- xmpp: calls to `Serve` no longer return `io.EOF` on success


### Fixed

- examples/echobot: calling `Send` from within the handler resulted in deadlock
- xmpp: closing the input stream was racy, resulting in invalid XML


## v0.12.0

### Breaking

- dial: moved network dialing types and functions into new package.
- dial: use underlying net.Dialer's DNS Resolver in Dialer.
- stanza: change API of `WrapIQ` and `WrapPresence` to not abuse pointers
- xmpp: add new `SendIQ` API and remove response from `Send` and `SendElement`
- xmpp: new API for writing custom tokens to a session

### Fixed

- xmpp: let `Session.Close` operate concurrently with `SendElement` et al.
