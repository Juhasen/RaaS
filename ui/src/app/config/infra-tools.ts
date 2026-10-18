export interface InfraTool {
  id: string;
  name: string;
  description: string;
  url: string;
}

// Local cluster NodePorts, see k8s/infra/admin-tools. Change here for another cluster.
export const INFRA_TOOLS: readonly InfraTool[] = [
  { id: 'minio', name: 'MinIO', description: 'Object storage console', url: 'http://localhost:30901' },
  { id: 'postgres', name: 'Postgres (Adminer)', description: 'Browse and query the database', url: 'http://localhost:30090' },
  { id: 'kafka', name: 'Kafka (Redpanda Console)', description: 'Topics, messages and consumer groups', url: 'http://localhost:30808' },
  { id: 'grafana', name: 'Grafana', description: 'Metrics and log dashboards', url: 'http://localhost:30300' }
];
