# Contract: Infrastructure Port Mappings

1. PostgreSQL: `5432` on `postgres.default.svc.cluster.local`
2. MongoDB: `27017` on `mongodb.default.svc.cluster.local`
3. Redis: `6379` on `redis.default.svc.cluster.local`
4. Kafka: `9092` on `kafka.default.svc.cluster.local`

## Language Tracks Local Dev Ports (`NodePort`)

**Go (`k8s/apps/go`)**
- listing-service: `<NodeIP>:30001` mapping to Container `8080`
- booking-service: `<NodeIP>:30002` mapping to Container `8080`
- media-service: `<NodeIP>:30003` mapping to Container `8080`

**Java (`k8s/apps/java`)**
- payment-service: `<NodeIP>:30011` mapping to Container `8080`
- review-service: `<NodeIP>:30012` mapping to Container `8080`
- favorites-service: `<NodeIP>:30013` mapping to Container `8080`

**Python (`k8s/apps/python`)**
- user-service: `<NodeIP>:30021` mapping to Container `8080`
- notification-service: `<NodeIP>:30022` mapping to Container `8080`
- analytics-service: `<NodeIP>:30023` mapping to Container `8080`