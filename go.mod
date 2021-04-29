module maunium.net/go/mautrix-whatsapp

go 1.14

require (
	github.com/Rhymen/go-whatsapp v0.1.0
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.9.0
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/prometheus/client_golang v1.9.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	gopkg.in/yaml.v2 v2.3.0
	maunium.net/go/mauflag v1.0.0
	maunium.net/go/maulogger/v2 v2.2.4
	maunium.net/go/mautrix v0.9.8
)

replace github.com/Rhymen/go-whatsapp => github.com/pidongqianqian/go-whatsapp v0.4.6-0.20210429085121-f27237dc7eec

replace maunium.net/go/mautrix => github.com/pidongqianqian/mautrix-go v0.9.9-0.20210428105727-2cf487e9936b
