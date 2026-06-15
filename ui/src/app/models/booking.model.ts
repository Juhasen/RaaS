export interface Booking {
  id?: string;
  listing_id: string;
  guest_id: string;
  start_date: string;
  end_date: string;
  total_price: number;
  status?: string;
  created_at?: string;
}
