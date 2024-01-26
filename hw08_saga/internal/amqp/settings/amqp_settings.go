package amqp_settings

const ExchangeOrder = "shop_exchange_order"
const QueueOrderPayService = "shop_queue_order_for_pay_service"
const QueueOrderStockService = "shop_queue_order_for_stock_service"
const QueueOrderDeliveryService = "shop_queue_order_for_delivery_service"

const RoutingKeyPayService = "pay"
const RoutingKeyStockService = "stock"
const RoutingKeyDeliveryService = "delivery"

const ExchangeStatus = "shop_exchange_status"
const QueueStatusOrderService = "shop_queue_status_for_order_service"
const QueueStatusPayService = "shop_queue_status_for_pay_service"
const QueueStatusStockService = "shop_queue_status_for_stock_service"
const QueueStatusDeliveryService = "shop_queue_status_for_delivery_service"
