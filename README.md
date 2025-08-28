

1. Transport Layer (UDP-Socket Handling)

    Non-blocking UDP-Socket

    Empfangs- und Sendepuffer

    Event-Loop (z. B. mit select, epoll, tokio, etc.)

2. Packet Framing

3.Prozessablauf: Minimaler QUIC-ähnlicher Flow

    Client sendet Initial-Paket → Enthält Session-ID, Version, optional ein Token oder Public Key.

    Server antwortet mit Ack oder Retry → Bei Retry: Token zurücksenden, ggf. Proof-of-Work verlangen.

    Handshake (optional) → Austausch von Schlüsseln, Capabilities, Versionen.

    Session aktiv → Datenübertragung über Streams, mit Sequenznummern und optionalen Acks.

    Session-Ende → FIN-Flag oder Timeout, ggf. Session-Ticket für Resume.



Verbindungsaufbau	Kein Aufbau, sofort sendbar	3‑Wege‑Handshake	0‑RTT möglich, TLS 1.3 integriert
Zuverlässigkeit	Keine	Vollständig (ACKs, Retransmit)	Vollständig, aber im User‑Space
Reihenfolgegarantie	Keine	Ja	Ja, pro Stream
Multiplexing	Manuell	Nicht nativ	Native Streams ohne Head‑of‑Line Blocking
Fehlererkennung	UDP‑Checksumme	Checksumme + ACK/NACK	AEAD‑Tag + Retry‑Mechanismen
Staukontrolle	Keine	TCP Reno/Cubic etc.	BBR/Cubic im User‑Space
Verschlüsselung	Optional (z. B. DTLS)	Extern via TLS	Integriert (TLS 1.3)
Headerstruktur	Minimal, fix	Fix, aber komplex	Zwei Headertypen, teils verschlüsselt
NAT‑Traversal	Manuell (STUN/ICE)	NAT‑freundlich	NAT‑freundlich, Connection‑ID
Fragmentierung	Risiko bei > MTU	Automatisch	App‑Layer Segmentierung
Session‑Management	Manuell	Implizit über Socket	Explizit mit Connection‑ID


Transportfunktionen über udp

    Zuverlässigkeit (optional): Stop‑and‑Wait, Sliding Window (Go‑Back‑N, Selective Repeat), selektive Acks, Retransmit-Timeouts, Sack-Bitmap.

    Reihenfolgegarantie (optional): Reordering-Buffer je Stream, In‑Order‑Delivery Flag.

    Fluss- & Staukontrolle: Empfangsfenster, Token-/Leaky‑Bucket, Rate Limiting, rudimentäre Congestion Control (langsam starten, bei Verlusten drosseln).

    FEC/Redundanz (optional): XOR/RS/Raptor für Verlusttoleranz statt Retransmits bei Live-Streaming.

    Heartbeats & Keepalive: Periodische Pings, RTT-/Jitter‑Schätzung, Session-Timeouts.

    Fragmentierung vermeiden: Payload < Path MTU halten (z. B. ≤ 1200 Byte), PLPMTUD oder Probe-Pakete; keine IP‑Fragmentierung.

    NAT‑Traversal (bei P2P): STUN/ICE-Unterstützung, regelmäßige Keepalives, Port-Pinning.

Sicherheit und sitzungsmanagement

    Verschlüsselung: DTLS oder eigenes Schema mit AEAD (z. B. ChaCha20‑Poly1305/AES‑GCM); Nonces/IV per Sequenznummer ableiten.

    Authentizität/Integrität: HMAC/AEAD‑Tag über Header+Payload; Schutz vor Spoofing.

    Schlüsseltausch: PSK, ECDH‑Handshake, Session-Resumption; Re‑Keying bei Langläufern.

    Replay‑Schutz: Strict Monotonic Sequenzen, Window‑Check, Tokens.

    DoS-/Amplification‑Schutz: Stateless Retry Token, Rate‑Limiting pro 5‑Tuple, Cookie/Proof‑of‑Work vor großen Antworten.

    Sitzungszustand: Zustandsmaschine (Init → Handshake → Active → Closing), Session-Tickets, Version-Negotiation.

System- und I/O-architektur

    Socket‑Ebene: Non‑blocking UDP, Event‑Loop (epoll/kqueue/IOCP), Puffergrößen (SO_RCVBUF/SO_SNDBUF), Timestamping.

    Threading/Concurrency: Single‑threaded Reactor vs. Worker‑Pools; Lock‑freie Queues für Ingress/Egress.

    Paket‑Pipelines: Receive → Parse → Validate → Decrypt → Route (Session/Stream) → Reorder → App; umgekehrt beim Senden.

    Serialisierung: Einfaches Binary‑Schema, TLV, oder CBOR/Protobuf/FlatBuffers; Endianness klar definieren (meist Network Byte Order).

    Timer: Hochauflösende Timer für Retransmits, Heartbeats, TLP/Probe‑Pakete, Idle‑Timeouts.

    Priorisierung: Separate Queues für Control vs. Data, Drop‑Policy bei Druck (zuerst unkritische Pakete).



