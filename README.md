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
– Add retries on requests 
– Validate addresses
– Pay to not get rate limited on solana
– Graceful shutdown

TODO
- Check direciton of chan in signautre of funciton <-chan or chan<-
- put all the const on top of file at least like ticker time
– add database to put the last block done
- Add Bitcoin watcher
