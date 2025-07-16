# Backend Interview Crypto

1. Copy `.env.example` as `.env` and fill it.

2. Start Kafka
```bash
docker-compose up -d
```

3. Start the service
```bash
go run cmd/main.go
```

4. See transactions in Kafka
```bash
docker-compose exec kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic transactions --from-beginning
```

Check transactions on explorers: [Solana](https://solana.fm/?cluster=mainnet-alpha), [Ethereum](https://etherscan.io/)

Improvements:
– Add Bitcoin watcher
– Add retries on requests 
– Validate addresses
– Pay to not get rate limited on solana
– Graceful shutdown
- Add a database to restart on the current block stored

Todo:
- Retry the code if all good, update api variables.
- redo the README
- check everything a last time