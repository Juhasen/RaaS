export interface Listing {
  id?: string;
  host_id: string;
  title: string;
  description: string;
  price_per_day: number;
  location_id: string;
  location_label: string;
  status?: string;
  media_urls?: string[] | null;
  created_at?: string;
}

export interface PaginatedListings {
  data: Listing[];
  total: number;
  page: number;
  limit: number;
}
