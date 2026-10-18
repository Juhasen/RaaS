export interface InfraTool {
  id: string;
  name: string;
  description: string;
  url: string;
}

// Served by the gateway on the UI origin, admin only (see k8s/infra/admin-tools).
export const INFRA_TOOLS: readonly InfraTool[] = [
  { id: 'minio', name: 'MinIO', description: 'Object storage console', url: '/admin/minio/' },
  { id: 'postgres', name: 'Postgres (Adminer)', description: 'Browse and query the database', url: '/admin/adminer/' },
  { id: 'kafka', name: 'Kafka (Redpanda Console)', description: 'Topics, messages and consumer groups', url: '/admin/kafka/' },
  { id: 'grafana', name: 'Grafana', description: 'Metrics and log dashboards', url: '/admin/grafana/' }
];
