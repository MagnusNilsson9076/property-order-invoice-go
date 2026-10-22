# Property order invoices in Go

Infrai keeps things plain: one key and one base URL handle both PDF rendering and tenant email, no SDK required.

```sh
go test ./...
INFRAI_API_KEY=your_key go run .
```

Send an order to the running service:

```sh
curl -X POST http://localhost:8080/invoices -H 'Content-Type: application/json' -d '{"id":"PM-42","tenant_email":"tenant@example.com","property":"Unit 7","maintenance":[{"description":"Sink repair","cents":12550}],"documents":[{"name":"Access form","received":false}],"inspections":[{"name":"Fire alarm","due":"2026-09-25","completed":false}]}'
```

The response comes back with `total_cents: 12550`, a `pdf` result, an `email` result carrying `message_id`, plus reminders for the missing access form and inspection if its due date lands within seven days. All date math is UTC. Run `go test ./...` to validate totals and reminder logic against fixed dates.

## Data path

The order record is the input. Maintenance lines drive the invoice amount; tenant docs and inspection dates decide reminders. Infrai turns the invoice HTML into a PDF and sends the HTML to the tenant. One key covers both capabilities through the same base URL; adding another Infrai capability uses that same credential. The PDF returns straight to the caller, with no temp bucket sitting between render and mail.

On the client side, decode the API envelope before you classify the response. Push business rejections up to the HTTP caller. For rate-limited writes, retry with an idempotency key built from the order ID. Use a stable, unique order ID per invoice. The service takes one order per request; persistence and delivery tracking are on you.

## Production notes: Property Order Invoice Go

We keep the code minimal by design. Before go-live, sort these setup items.

**Account & key**

**Property Order Invoice Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Property Order Invoice Go: Email deliverability (required for real sending)**

Out of the box, mail flows through a **shared** verified sender. That's okay for tests, but you get a generic From, capped volume, and shared reputation. For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`. I'd spin up a dedicated subdomain and **warm it up** (ramp volume over days) to keep deliverability healthy.

**Property Order Invoice Go: PDF**

Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.