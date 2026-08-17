package booking

import (
	"fmt"
	"os"
	"strconv"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

// 🔴 เพิ่ม bookingID เข้ามาใน Signature เพื่อเตรียมความพร้อมสำหรับ Webhook
func CreateStripeCheckout(bookingID uint, eventID uint, seatID uint, price float64, userID string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// ใช้ NEXT_PUBLIC_URL ตาม V1 และใส่ Fallback ป้องกันค่าว่างตาม V2
	domain := os.Getenv("NEXT_PUBLIC_URL")
	if domain == "" {
		domain = "http://localhost:3000"
	}

	params := &stripe.CheckoutSessionParams{
		// 🔴 รวมช่องทางการชำระเงินทั้ง Card และ PromptPay
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "promptpay"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("thb"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						// แสดงรายละเอียดเลขที่นั่งให้ลูกค้าเห็นชัดเจน
						Name: stripe.String(fmt.Sprintf("Concert Ticket - Seat %d", seatID)),
					},
					UnitAmount: stripe.Int64(int64(price * 100)), // Stripe ใช้หน่วยสตางค์ (คูณ 100)
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		
		// 🔴 ส่ง Metadata ไปครบถ้วน โดยมี booking_id เป็น Primary Identifier สำหรับ Webhook
		Metadata: map[string]string{
			"booking_id": strconv.FormatUint(uint64(bookingID), 10),
			"event_id":   strconv.FormatUint(uint64(eventID), 10),
			"seat_id":    strconv.FormatUint(uint64(seatID), 10),
			"user_id":    userID,
		},

		// คงรูปแบบ URL กลับไปที่หน้า Event เดิมตาม V1 เพื่อ UX ที่ดีกว่า
		SuccessURL: stripe.String(domain + "/events/" + fmt.Sprintf("%d", eventID) + "?success=true"),
		CancelURL:  stripe.String(domain + "/events/" + fmt.Sprintf("%d", eventID) + "?canceled=true"),
	}

	s, err := session.New(params)
	if err != nil {
		return "", err
	}

	return s.URL, nil
}