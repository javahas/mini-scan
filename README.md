# Mini-Scan

Hello!

As you've heard by now, Censys scans the internet at an incredible scale. Processing the results necessitates scaling horizontally across thousands of machines. One key aspect of our architecture is the use of distributed queues to pass data between machines.

---

The `docker-compose.yml` file sets up a toy example of a scanner. It spins up a Google Pub/Sub emulator, creates a topic and subscription, and publishes scan results to the topic. It can be run via `docker compose up`.

Your job is to build the data processing side. It should:

1. Pull scan results from the subscription `scan-sub`.
2. Maintain an up-to-date record of each unique `(ip, port, service)`. This should contain when the service was last scanned and a string containing the service's response.

> **_NOTE_**
> The scanner can publish data in two formats, shown below. In both of the following examples, the service response should be stored as: `"hello world"`.
>
> ```javascript
> {
>   // ...
>   "data_version": 1,
>   "data": {
>     "response_bytes_utf8": "aGVsbG8gd29ybGQ="
>   }
> }
>
> {
>   // ...
>   "data_version": 2,
>   "data": {
>     "response_str": "hello world"
>   }
> }
> ```

Your processing application should be able to be scaled horizontally, but this isn't something you need to actually do. The processing application should use `at-least-once` semantics where ever applicable.

You may write this in any languages you choose, but Go would be preferred.

You may use any data store of your choosing, with `sqlite` being one example. Like our own code, we expect the code structure to make it easy to switch data stores.

Please note that Google Pub/Sub is best effort ordering and we want to keep the latest scan. While the example scanner does not publish scans at a rate where this would be an issue, we expect the application to be able to handle extreme out of orderness. Consider what would happen if the application received a scan that is 24 hours old.

cmd/scanner/main.go should not be modified

---


## Solution
- `cmd/processor` implements a Pub/Sub consumer that decodes scan messages and keeps the latest scan per `(ip, port, service)` based on `timestamp`.
- `pkg/storage` provides a `Store` interface with file (`FileStore`) and SQLite (`SQLiteStore`) implementations to make storage swapping.


## How to Run

1. Start the emulator, topic/subscription creation, and scanner:
   ```bash
   docker compose up
   ```
2. In another shell, run the processor:
   - For File storage:

      ```bash
         PUBSUB_EMULATOR_HOST=localhost:8085 go run ./cmd/processor --store-type=file --store-path=data/file_recs.json
      ```
   - For SQL Lite storage:

      ```bash
       PUBSUB_EMULATOR_HOST=localhost:8085 go run ./cmd/processor --store-type=sqlite --store-path=data/sqllite_recs.db
      ```      

3. Inspect the stored results at `data/file_recs.json` or `data/sqllite_recs.db`.
