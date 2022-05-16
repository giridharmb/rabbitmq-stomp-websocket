## DB Setup

#### Bash ENV Exports

```bash
export MQUSER="..."
export MQPASS="..."
export MQHOST="..."

export TEST_USER="..."
export TEST_PASS="..."
export TEST_DB="..."
export TEST_PGSQL_HOST="..."
```

#### Build The Binary

```bash
go build -o mq_ws
```

#### Running HTTP WebServer (API)

```bash
./mq_ws -o api
```

#### Pushing Messages on Exchange (Broadcast To All)

```bash
./mq_ws -o push_messages_on_exchange
```

