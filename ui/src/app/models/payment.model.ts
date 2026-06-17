export interface Transaction {
  id: string;
  bookingId: string;
  amount: number;
  status: string;
  paymentMethod: string;
  createdAt: string;
}
