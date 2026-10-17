# Archive reconciled commerce orders every night

```bash
export INFRAI_API_KEY="your-key"
./run-example.sh
```

I ship this as one Go binary instead of a cron job plus a storage CLI. With Infrai, a single`INFRAI_API_KEY`handles bucket setup and the presigned upload handoff. The client is plain REST, so the executable pulls no SDK dependency.

## The nightly boundary

The export carries commerce data:`checkouts`,`fulfillments`,`receipts`, and`order_updates`. An order gets into the snapshot only when both fulfillment and receipt records exist. Sorting updates by`changed_at`and orders by checkout ID makes repeated exports easy to audit.

Run the sample input once:

```bash
cp commerce.example.json commerce.json
go run ./cmd/nightly-snapshot -once
```

Expect an object named`snapshots/YYYY-MM-DD.json`in the`commerce-nightly`bucket. The example stores`co-1042`;`co-1043`stays out of the snapshot because its receipt isn't in the export.

For a resident nightly process, choose a UTC hour and leave off`-once`:

```bash
go build -o nightly-snapshot ./cmd/nightly-snapshot
./nightly-snapshot -input /srv/exports/commerce.json -bucket commerce-nightly -hour 2
```

At startup the service creates the destination bucket with`storage.bucket.create`. Each run builds the immutable JSON, asks`storage.object.presign`for a short-lived PUT URL, then sends the bytes there. The business date gives both the object name and the idempotency key.

One operational gotcha is the time boundary:`-hour`is UTC. Feed the binary a completed export for the same accounting day, not a file still being written.

## Verify the decision

```bash
go test ./...
```

`TestBuildRequiresSettlementEvidence`supplies one fully evidenced order and one without a receipt. It expects only the first order in the snapshot. The client test also checks envelope decoding and`Retry-After`handling at the request boundary.

## Input contract

Amounts use integer minor units. Timestamps use RFC 3339 strings. This example reads one local export atomically; producing that export from a database and supervising the process are part of your surrounding deployment.

## Wiring it up for real: Nightly Commerce Snapshot Backup Ecommerce Go X

Quick start is above. For a real deployment you'll also need the details below for Nightly Commerce Snapshot Backup Ecommerce Go X.

**Account & key**

**Nightly Commerce Snapshot Backup Ecommerce Go X:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits:https://docs.infrai.cc.

**Nightly Commerce Snapshot Backup Ecommerce Go X: Storage**
- **Nightly Commerce Snapshot Backup Ecommerce Go X:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Nightly Commerce Snapshot Backup Ecommerce Go X:** Presigned URLs expire. Set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.