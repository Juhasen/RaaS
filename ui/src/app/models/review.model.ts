export interface Review {
  id?: string;
  bookingId: string;
  reviewerId: string;
  listingId: string;
  rating: number;
  comment: string;
  createdAt?: string;
}
