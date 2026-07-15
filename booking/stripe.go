package booking

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func CreateStripeCheckout(eventID uint, seatID uint, price float64, userID string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	domain := os.Getenv("NEXT_PUBLIC_URL")

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("thb"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Concert Ticket - Seat %d", seatID)),
					},
					UnitAmount: stripe.Int64(int64(price * 100)), // Stripe ใช้หน่วยสตางค์ (คูณ 100)
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		// ส่ง metadata ไปด้วย เพื่อให้ Webhook รู้ว่าจ่ายค่าที่นั่งไหน
		Metadata: map[string]string{
			"user_id":  userID,
			"event_id": fmt.Sprintf("%d", eventID),
			"seat_id":  fmt.Sprintf("%d", seatID),
		},
	
		SuccessURL: stripe.String(domain + "/events/" + fmt.Sprintf("%d", eventID) + "?success=true"),
		CancelURL:  stripe.String(domain + "/events/" + fmt.Sprintf("%d", eventID) + "?canceled=true"),
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}
	return s.URL, nil
}