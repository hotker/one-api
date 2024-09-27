package stripe

import (
	"encoding/json"
	"fmt"
	"math"
	"one-api/payment/types"
	"strconv"

	sysconfig "one-api/common/config"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72/webhook"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
)

// Stripe 结构体实现支付接口
type Stripe struct{}

// Name 返回支付方式名称
func (e *Stripe) Name() string {
	return "Stripe"
}

// Pay 处理支付请求
func (e *Stripe) Pay(config *types.PayConfig, gatewayConfig string) (*types.PayRequest, error) {
	var stripeConfig StripeConfig
	// 使用 json.Unmarshal 解析 JSON 字符串到结构体
	err := json.Unmarshal([]byte(gatewayConfig), &stripeConfig)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	sc := &client.API{}
	sc.Init(stripeConfig.SecretKey, nil)
	currency := stripe.String("USD")
	if config.Currency == "CNY" {
		currency = stripe.String("CNY")
	}

	result, err := sc.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(config.ReturnURL),
		ClientReferenceID: stripe.String(config.TradeNo),

		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: currency,
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(sysconfig.SystemName + "-Token充值:" + strconv.FormatFloat(config.Money, 'f', 0, 64) + " " + string(config.Currency)),
					},
					UnitAmount: stripe.Int64(int64(math.Round(config.Money * 100))),
				},
				Quantity: stripe.Int64(1),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	// 构造支付请求
	payRequest := &types.PayRequest{
		Type: 1,
		Data: types.PayRequestData{
			URL: result.URL,
			Params: map[string]interface{}{
				"tradeNo": config.TradeNo,
				"linkId":  result.ID,
			},
		},
	}

	return payRequest, nil
}

// HandleCallback 处理支付回调
func (e *Stripe) HandleCallback(c *gin.Context, gatewayConfig string) (*types.PayNotify, error) {
	body, err := c.GetRawData()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	var stripeConfig StripeConfig

	if err := json.Unmarshal([]byte(gatewayConfig), &stripeConfig); err != nil {
		return nil, fmt.Errorf("failed to parse gateway config: %v", err)
	}

	sc := &client.API{}

	sc.Init(stripeConfig.SecretKey, nil)
	stripeSignature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, stripeSignature, stripeConfig.WebhookSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to verify webhook: %v", err)
	}

	// 处理事件
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return nil, fmt.Errorf("failed to parse session data: %v", err)
		}

		// 获取订单号
		orderID := session.ClientReferenceID

		// 构造 PayNotify
		payNotify := &types.PayNotify{
			TradeNo:   orderID,
			GatewayNo: session.PaymentIntent.ID,
		}

		return payNotify, nil
	default:
		return nil, nil
	}
}