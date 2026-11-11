export interface NotificationItem {
  id: string;
  user_id: string;
  message: string;
  notification_type: string;
  status: string;
  created_at: string;
}

export interface NotificationListResponse {
  total: number;
  count: number;
  items: NotificationItem[];
}
